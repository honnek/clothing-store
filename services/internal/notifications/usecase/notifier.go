package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/honnek/lumewear-shop/services/internal/notifications/domain"
	"github.com/honnek/lumewear-shop/services/internal/platform/kafka"
	platformotel "github.com/honnek/lumewear-shop/services/internal/platform/otel"
)

var (
	sent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "notifications_sent_total",
		Help: "Отправленные письма о новых заказах.",
	})

	skipped = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "notifications_skipped_total",
		Help: "Сообщения, по которым письмо не отправлялось.",
	}, []string{"reason"})
)

func init() {
	prometheus.MustRegister(sent, skipped)
}

// Mailer — порт почты. Реализация ходит в SMTP, в тестах — фейк.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Dedup отвечает на вопрос «это событие уже обрабатывали?». Kafka даёт at-least-once,
// а второе письмо об одном заказе покупателю не нужно. Отметка ставится до отправки
// (иначе две реплики воркера отправят по письму на одно событие), поэтому нужен и
// Forget — снять отметку, если письмо всё-таки не ушло.
type Dedup interface {
	FirstSeen(ctx context.Context, key string) (bool, error)
	Forget(ctx context.Context, key string) error
}

type Notifier struct {
	mailer Mailer
	dedup  Dedup
	log    *slog.Logger
	to     string
}

func New(mailer Mailer, dedup Dedup, log *slog.Logger, to string) *Notifier {
	return &Notifier{mailer: mailer, dedup: dedup, log: log, to: to}
}

// Handle обрабатывает одно сообщение из топика.
//
// Ошибку возвращаем только на временных сбоях (почта или redis не ответили) — тогда
// воркер встанет и после перезапуска прочитает сообщение снова. Битый payload ошибкой
// не считаем: он не починится сам и заклинит раздел навсегда, поэтому логируем и идём
// дальше.
func (n *Notifier) Handle(ctx context.Context, msg kafka.Message) error {
	ctx, span := otel.Tracer("notifications").Start(
		platformotel.FromCarrier(ctx, msg.Headers),
		"notifications order.created",
		trace.WithSpanKind(trace.SpanKindConsumer),
	)
	defer span.End()

	var event domain.OrderCreated
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		skipped.WithLabelValues("malformed").Inc()
		n.log.ErrorContext(ctx, "malformed order.created",
			slog.String("key", msg.Key), slog.Any("error", err))
		return nil
	}
	span.SetAttributes(attribute.Int("order.id", int(event.OrderID)))

	key := fmt.Sprintf("notify:order.created:%d", event.OrderID)
	first, err := n.dedup.FirstSeen(ctx, key)
	if err != nil {
		return fmt.Errorf("dedup order %d: %w", event.OrderID, err)
	}
	if !first {
		skipped.WithLabelValues("duplicate").Inc()
		n.log.InfoContext(ctx, "order.created already notified", slog.Int("order_id", int(event.OrderID)))
		return nil
	}

	subject := fmt.Sprintf("Заказ №%d принят", event.OrderID)
	if err := n.mailer.Send(ctx, n.to, subject, letter(event)); err != nil {
		if ferr := n.dedup.Forget(ctx, key); ferr != nil {
			n.log.ErrorContext(ctx, "dedup mark left behind after failed send",
				slog.String("key", key), slog.Any("error", ferr))
		}
		return fmt.Errorf("send mail for order %d: %w", event.OrderID, err)
	}

	sent.Inc()
	n.log.InfoContext(ctx, "order.created notified",
		slog.Int("order_id", int(event.OrderID)), slog.String("total", event.Total))
	return nil
}

func letter(e domain.OrderCreated) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Заказ №%d от %s\n\n", e.OrderID, e.CreatedAt.Format("02.01.2006 15:04"))
	for _, it := range e.Items {
		fmt.Fprintf(&b, "%s — %d × %s = %s\n", it.Title, it.Quantity, it.UnitPrice, it.LineTotal)
	}
	fmt.Fprintf(&b, "\nИтого: %s\n", e.Total)
	if e.OwnerID != nil {
		fmt.Fprintf(&b, "Покупатель: #%d\n", *e.OwnerID)
	}
	return b.String()
}
