// Package outbox содержит relay: он вычитывает события, которые checkout положил в
// таблицу outbox, и публикует их в брокер. Отдельный процесс публикации нужен потому,
// что запись в Kafka нельзя внести в ту же транзакцию, что и заказ.
package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/honnek/lumewear-shop/services/internal/order/domain"
	"github.com/honnek/lumewear-shop/services/internal/platform/kafka"
	platformotel "github.com/honnek/lumewear-shop/services/internal/platform/otel"
)

var (
	published = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "outbox_events_published_total",
		Help: "События outbox, доехавшие до брокера.",
	}, []string{"event_type"})

	failures = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_relay_failures_total",
		Help: "Неудачные проходы relay (брокер или БД не ответили).",
	})

	pending = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_pending_events",
		Help: "Неопубликованные события в outbox — лаг доставки.",
	})
)

func init() {
	prometheus.MustRegister(published, failures, pending)
}

// Store — то, что relay ждёт от хранилища. Границу транзакции держит адаптер:
// publish вызывается внутри неё, до коммита пометки published_at.
// Пока пачка не отправлена, транзакция с FOR UPDATE держится открытой, а брокер
// в недоступности умеет ждать бесконечно — отсюда жёсткий потолок на проход.
const drainTimeout = 15 * time.Second

type Store interface {
	DrainOutbox(ctx context.Context, batch int32, publish func(context.Context, []domain.OutboxEvent) error) (int, error)
	PendingOutbox(ctx context.Context) (int64, error)
}

type Publisher interface {
	Publish(ctx context.Context, msg kafka.Message) error
}

type Relay struct {
	store Store
	pub   Publisher
	log   *slog.Logger

	topic string
	batch int32
	every time.Duration
}

func NewRelay(store Store, pub Publisher, log *slog.Logger, topic string, batch int32, every time.Duration) *Relay {
	return &Relay{store: store, pub: pub, log: log, topic: topic, batch: batch, every: every}
}

// Run крутит цикл до отмены ctx. Полная пачка означает, что в очереди осталось ещё —
// тогда идём за следующей сразу, не досиживая тик, иначе всплеск заказов будет
// разгребаться со скоростью одного тика на пачку.
func (r *Relay) Run(ctx context.Context) error {
	t := time.NewTicker(r.every)
	defer t.Stop()

	for {
		for {
			n, err := r.drain(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				failures.Inc()
				r.log.ErrorContext(ctx, "outbox relay", slog.Any("error", err))
				break
			}
			if n < int(r.batch) {
				break
			}
		}

		if left, err := r.store.PendingOutbox(ctx); err == nil {
			pending.Set(float64(left))
		}

		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}

func (r *Relay) drain(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, drainTimeout)
	defer cancel()

	return r.store.DrainOutbox(ctx, r.batch, func(ctx context.Context, events []domain.OutboxEvent) error {
		for _, e := range events {
			if err := r.publish(ctx, e); err != nil {
				return fmt.Errorf("publish outbox %d: %w", e.ID, err)
			}
		}
		return nil
	})
}

func (r *Relay) publish(ctx context.Context, e domain.OutboxEvent) error {
	// Родитель спана — трейс checkout-а, сохранённый в строке: сам checkout к этому
	// моменту давно ответил клиенту, и in-process контекста здесь уже нет.
	parent := platformotel.FromCarrier(ctx, map[string]string{"traceparent": e.Traceparent})
	spanCtx, span := otel.Tracer("outbox").Start(parent, "outbox publish "+e.Type,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.destination.name", r.topic),
			attribute.String("event.type", e.Type),
		))
	defer span.End()

	err := r.pub.Publish(spanCtx, kafka.Message{
		Topic: r.topic,
		// Ключ — id заказа: события одного заказа лягут в один раздел и придут
		// потребителю по порядку.
		Key:     e.AggregateID,
		Value:   e.Payload,
		Headers: platformotel.Carrier(spanCtx),
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	published.WithLabelValues(e.Type).Inc()
	return nil
}
