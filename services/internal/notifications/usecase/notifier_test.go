package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/honnek/lumewear-shop/services/internal/platform/kafka"
)

type fakeMailer struct {
	sent []string
	err  error
}

func (m *fakeMailer) Send(_ context.Context, _, subject, body string) error {
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, subject+"\n"+body)
	return nil
}

type fakeDedup struct {
	seen map[string]bool
	err  error
}

func (d *fakeDedup) FirstSeen(_ context.Context, key string) (bool, error) {
	if d.err != nil {
		return false, d.err
	}
	if d.seen[key] {
		return false, nil
	}
	if d.seen == nil {
		d.seen = map[string]bool{}
	}
	d.seen[key] = true
	return true, nil
}

func (d *fakeDedup) Forget(_ context.Context, key string) error {
	delete(d.seen, key)
	return nil
}

func newNotifier(m Mailer, d Dedup) *Notifier {
	return New(m, d, slog.New(slog.NewTextHandler(io.Discard, nil)), "orders@shop.local")
}

func message(payload string) kafka.Message {
	return kafka.Message{Topic: "order.created", Key: "7", Value: []byte(payload)}
}

const orderCreated = `{"order_id":7,"total":"49.98","created_at":"2026-08-31T10:00:00Z",
	"items":[{"title":"Кепка","unit_price":"24.99","quantity":2,"line_total":"49.98"}]}`

func TestHandleSendsLetter(t *testing.T) {
	mailer := &fakeMailer{}

	if err := newNotifier(mailer, &fakeDedup{}).Handle(context.Background(), message(orderCreated)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(mailer.sent) != 1 {
		t.Fatalf("sent %d letters, want 1", len(mailer.sent))
	}

	letter := mailer.sent[0]
	for _, want := range []string{"Заказ №7", "Кепка", "49.98"} {
		if !strings.Contains(letter, want) {
			t.Errorf("letter has no %q:\n%s", want, letter)
		}
	}
}

// Повторная доставка того же события — штатная работа Kafka, второе письмо покупателю
// не уходит.
func TestHandleSkipsDuplicate(t *testing.T) {
	mailer := &fakeMailer{}
	n := newNotifier(mailer, &fakeDedup{})

	for i := 0; i < 3; i++ {
		if err := n.Handle(context.Background(), message(orderCreated)); err != nil {
			t.Fatalf("handle #%d: %v", i, err)
		}
	}
	if len(mailer.sent) != 1 {
		t.Fatalf("sent %d letters, want 1", len(mailer.sent))
	}
}

// Битое сообщение не должно вставать поперёк раздела: логируем и признаём обработанным.
func TestHandleSwallowsMalformedPayload(t *testing.T) {
	mailer := &fakeMailer{}

	if err := newNotifier(mailer, &fakeDedup{}).Handle(context.Background(), message("{oops")); err != nil {
		t.Fatalf("handle = %v, want nil", err)
	}
	if len(mailer.sent) != 0 {
		t.Fatalf("sent %d letters, want 0", len(mailer.sent))
	}
}

func TestHandleReturnsErrorOnTransientFailure(t *testing.T) {
	down := errors.New("smtp unavailable")

	err := newNotifier(&fakeMailer{err: down}, &fakeDedup{}).Handle(context.Background(), message(orderCreated))
	if !errors.Is(err, down) {
		t.Fatalf("err = %v, want %v", err, down)
	}
}

// Упавшая отправка не должна съесть событие: отметка снимается, и повтор доходит
// до почты.
func TestHandleRetriesAfterFailedSend(t *testing.T) {
	mailer := &fakeMailer{err: errors.New("smtp unavailable")}
	dedup := &fakeDedup{}
	n := newNotifier(mailer, dedup)

	if err := n.Handle(context.Background(), message(orderCreated)); err == nil {
		t.Fatal("handle = nil, want error")
	}

	mailer.err = nil
	if err := n.Handle(context.Background(), message(orderCreated)); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if len(mailer.sent) != 1 {
		t.Fatalf("sent %d letters, want 1", len(mailer.sent))
	}
}
