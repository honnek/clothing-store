//go:build integration

package kafka

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcredpanda "github.com/testcontainers/testcontainers-go/modules/redpanda"
)

func discard() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func brokers(t *testing.T) []string {
	t.Helper()
	ctx := context.Background()

	rp, err := tcredpanda.RunContainer(ctx,
		testcontainers.WithImage("docker.redpanda.com/redpandadata/redpanda:v23.3.11"),
		tcredpanda.WithAutoCreateTopics(),
	)
	if err != nil {
		t.Fatalf("start redpanda: %v", err)
	}
	t.Cleanup(func() { _ = rp.Terminate(ctx) })

	addr, err := rp.KafkaSeedBroker(ctx)
	if err != nil {
		t.Fatalf("seed broker: %v", err)
	}
	return []string{addr}
}

func TestProduceConsume(t *testing.T) {
	seeds := brokers(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	producer, err := NewProducer(seeds)
	if err != nil {
		t.Fatal(err)
	}
	defer producer.Close()

	sent := Message{
		Topic:   "order.created",
		Key:     "42",
		Value:   []byte(`{"order_id":42}`),
		Headers: map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
	}
	if err := producer.Publish(ctx, sent); err != nil {
		t.Fatalf("publish: %v", err)
	}

	consumer, err := NewConsumer(seeds, "test-group", discard(), "order.created")
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Close()

	got := make(chan Message, 1)
	runCtx, stop := context.WithCancel(ctx)
	defer stop()

	done := make(chan error, 1)
	go func() {
		done <- consumer.Run(runCtx, func(_ context.Context, msg Message) error {
			got <- msg
			return nil
		})
	}()

	select {
	case msg := <-got:
		if msg.Key != sent.Key || string(msg.Value) != string(sent.Value) {
			t.Errorf("got %+v, want %+v", msg, sent)
		}
		if msg.Headers["traceparent"] != sent.Headers["traceparent"] {
			t.Errorf("traceparent = %q, want %q", msg.Headers["traceparent"], sent.Headers["traceparent"])
		}
	case err := <-done:
		t.Fatalf("consumer stopped early: %v", err)
	case <-ctx.Done():
		t.Fatal("message never arrived")
	}

	stop()
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}
}

// Обработчик, вернувший ошибку, останавливает цикл, не двинув смещение: поднятый
// заново потребитель той же группы получает то же сообщение.
func TestHandlerErrorStopsWithoutCommit(t *testing.T) {
	seeds := brokers(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	producer, err := NewProducer(seeds)
	if err != nil {
		t.Fatal(err)
	}
	defer producer.Close()

	if err := producer.Publish(ctx, Message{Topic: "retry.me", Key: "1", Value: []byte("x")}); err != nil {
		t.Fatal(err)
	}

	boom := errors.New("smtp down")
	first, err := NewConsumer(seeds, "retry-group", discard(), "retry.me")
	if err != nil {
		t.Fatal(err)
	}
	err = first.Run(ctx, func(context.Context, Message) error { return boom })
	first.Close()
	if !errors.Is(err, boom) {
		t.Fatalf("run = %v, want %v", err, boom)
	}

	second, err := NewConsumer(seeds, "retry-group", discard(), "retry.me")
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	redelivered := make(chan Message, 1)
	runCtx, stop := context.WithCancel(ctx)
	defer stop()
	go func() {
		_ = second.Run(runCtx, func(_ context.Context, msg Message) error {
			redelivered <- msg
			return nil
		})
	}()

	select {
	case msg := <-redelivered:
		if string(msg.Value) != "x" {
			t.Errorf("value = %q, want x", msg.Value)
		}
	case <-ctx.Done():
		t.Fatal("необработанное сообщение не пришло повторно")
	}
}
