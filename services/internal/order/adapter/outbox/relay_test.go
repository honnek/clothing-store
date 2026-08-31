package outbox

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/honnek/lumewear-shop/services/internal/order/domain"
	"github.com/honnek/lumewear-shop/services/internal/platform/kafka"
)

// fakeStore имитирует транзакцию адаптера: помечает события опубликованными только
// если publish не вернул ошибку.
type fakeStore struct {
	mu       sync.Mutex
	queue    []domain.OutboxEvent
	drained  int
	err      error
	deadline bool
}

func (s *fakeStore) DrainOutbox(ctx context.Context, batch int32, publish func(context.Context, []domain.OutboxEvent) error) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, s.deadline = ctx.Deadline()
	if s.err != nil {
		return 0, s.err
	}

	n := len(s.queue)
	if n > int(batch) {
		n = int(batch)
	}
	if n == 0 {
		return 0, nil
	}

	if err := publish(ctx, s.queue[:n]); err != nil {
		return 0, err
	}
	s.queue = s.queue[n:]
	s.drained += n
	return n, nil
}

func (s *fakeStore) PendingOutbox(context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return int64(len(s.queue)), nil
}

type fakePublisher struct {
	mu   sync.Mutex
	msgs []kafka.Message
	err  error
}

func (p *fakePublisher) Publish(_ context.Context, msg kafka.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return p.err
	}
	p.msgs = append(p.msgs, msg)
	return nil
}

func (p *fakePublisher) sent() []kafka.Message {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]kafka.Message(nil), p.msgs...)
}

func newRelay(store Store, pub Publisher, batch int32) *Relay {
	return NewRelay(store, pub, slog.New(slog.NewTextHandler(io.Discard, nil)),
		"order.created", batch, time.Millisecond)
}

func events(n int) []domain.OutboxEvent {
	out := make([]domain.OutboxEvent, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, domain.OutboxEvent{
			ID:          int64(i),
			AggregateID: "42",
			Type:        "order.created",
			Payload:     []byte(`{"order_id":42}`),
		})
	}
	return out
}

func TestRelayPublishesBatch(t *testing.T) {
	store := &fakeStore{queue: events(2)}
	pub := &fakePublisher{}

	n, err := newRelay(store, pub, 10).drain(context.Background())
	if err != nil {
		t.Fatalf("drain: %v", err)
	}
	if n != 2 {
		t.Fatalf("drained = %d, want 2", n)
	}

	msgs := pub.sent()
	if len(msgs) != 2 {
		t.Fatalf("published %d messages, want 2", len(msgs))
	}
	if msgs[0].Topic != "order.created" || msgs[0].Key != "42" {
		t.Errorf("message = %+v, want topic order.created and key 42", msgs[0])
	}
}

// Событие, которое брокер не принял, обязано остаться в очереди: relay без этого
// теряет заказ молча.
func TestRelayKeepsEventWhenPublishFails(t *testing.T) {
	store := &fakeStore{queue: events(1)}
	broken := errors.New("broker down")

	_, err := newRelay(store, &fakePublisher{err: broken}, 10).drain(context.Background())
	if !errors.Is(err, broken) {
		t.Fatalf("err = %v, want %v", err, broken)
	}
	if left, _ := store.PendingOutbox(context.Background()); left != 1 {
		t.Fatalf("pending = %d, want 1", left)
	}
}

// Очередь длиннее пачки разбирается за один тик, а не по пачке на тик.
func TestRelayDrainsQueueLongerThanBatch(t *testing.T) {
	store := &fakeStore{queue: events(5)}
	pub := &fakePublisher{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	relay := newRelay(store, pub, 2)
	go func() { done <- relay.Run(ctx) }()

	deadline := time.After(2 * time.Second)
	for {
		if len(pub.sent()) == 5 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("published %d of 5 events", len(pub.sent()))
		case <-time.After(5 * time.Millisecond):
		}
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}
}

// Молчащий брокер не должен держать открытой транзакцию с FOR UPDATE: у прохода
// обязан быть потолок по времени.
func TestRelayBoundsDrainTime(t *testing.T) {
	store := &fakeStore{queue: events(1)}

	if _, err := newRelay(store, &fakePublisher{}, 10).drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !store.deadline {
		t.Fatal("drain ran without a deadline")
	}
}

func TestRelaySurvivesStoreFailure(t *testing.T) {
	store := &fakeStore{err: errors.New("pg down")}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := newRelay(store, &fakePublisher{}, 10).Run(ctx); err != nil {
		t.Fatalf("run = %v, want nil: сбой БД не должен ронять сервис", err)
	}
}
