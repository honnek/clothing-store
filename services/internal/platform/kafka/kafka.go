// Package kafka — тонкая обёртка над franz-go: синхронный продюсер для outbox-relay
// и consumer-group для воркеров. Наружу отдаём Message с заголовками-строками, чтобы
// вызывающий код не тащил kgo-типы в свои порты.
package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Message — то, что кладём в топик и читаем из него. Headers несут W3C-контекст
// трейса, поэтому это map, а не пара полей.
type Message struct {
	Topic   string
	Key     string
	Value   []byte
	Headers map[string]string
}

type Producer struct {
	cl *kgo.Client
}

func NewProducer(brokers []string) (*Producer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		// Relay помечает событие опубликованным по факту ack от брокера, так что
		// ждём подтверждения от всех реплик — иначе пометка обгонит запись.
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerBatchMaxBytes(1<<20),
		// В dev-контуре топики никто заранее не заводит: разрешаем брокеру создать
		// order.created по первой публикации. В проде топик приезжает из инфраструктуры,
		// и флаг просто ничего не меняет.
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka producer: %w", err)
	}
	return &Producer{cl: cl}, nil
}

// Publish отправляет сообщение и ждёт подтверждения брокера.
func (p *Producer) Publish(ctx context.Context, msg Message) error {
	rec := &kgo.Record{
		Topic: msg.Topic,
		Key:   []byte(msg.Key),
		Value: msg.Value,
	}
	for k, v := range msg.Headers {
		rec.Headers = append(rec.Headers, kgo.RecordHeader{Key: k, Value: []byte(v)})
	}

	if err := p.cl.ProduceSync(ctx, rec).FirstErr(); err != nil {
		return fmt.Errorf("produce %s: %w", msg.Topic, err)
	}
	return nil
}

func (p *Producer) Close() { p.cl.Close() }

type Consumer struct {
	cl  *kgo.Client
	log *slog.Logger
}

func NewConsumer(brokers []string, group string, log *slog.Logger, topics ...string) (*Consumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topics...),
		kgo.ConsumerGroup(group),
		// Смещение двигаем сами после обработки: с автокоммитом упавший обработчик
		// потерял бы событие, а нам нужна ровно at-least-once доставка.
		kgo.DisableAutoCommit(),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: %w", err)
	}
	return &Consumer{cl: cl, log: log}, nil
}

// Run читает топики, пока не отменят ctx. Ошибка обработчика останавливает цикл без
// коммита — смещение остаётся на несъеденном сообщении, и после перезапуска воркер
// возьмёт его снова. Отсюда требование к обработчику: возвращать ошибку только на
// временных сбоях, а неразбираемое сообщение глотать самому, иначе цикл встанет намертво.
func (c *Consumer) Run(ctx context.Context, handle func(context.Context, Message) error) error {
	for {
		fetches := c.cl.PollFetches(ctx)
		if ctx.Err() != nil {
			return nil
		}
		// Ошибки выборки почти всегда временные (брокер перезапускается, лидер
		// переехал) — клиент переподключается сам, поэтому цикл не роняем: иначе
		// моргнувший брокер убивает воркер.
		for _, e := range fetches.Errors() {
			c.log.Warn("kafka fetch",
				slog.String("topic", e.Topic), slog.Int("partition", int(e.Partition)),
				slog.Any("error", e.Err))
		}

		if fetches.NumRecords() == 0 {
			// Пусто на фоне ошибок — брокера сейчас нет; не крутим цикл вхолостую,
			// иначе лог за минуту распухнет на тысячи строк.
			if len(fetches.Errors()) > 0 {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(time.Second):
				}
			}
			continue
		}

		var handleErr error
		fetches.EachRecord(func(rec *kgo.Record) {
			if handleErr != nil {
				return
			}
			handleErr = handle(ctx, toMessage(rec))
		})
		if handleErr != nil {
			return handleErr
		}

		if err := c.commit(ctx); err != nil {
			return err
		}
	}
}

// commit переживает отмену ctx: сообщения пачки уже обработаны, и терять их коммит
// на остановке сервиса — значит после рестарта отправить те же письма повторно.
func (c *Consumer) commit(ctx context.Context) error {
	commitCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if err := c.cl.CommitUncommittedOffsets(commitCtx); err != nil {
		return fmt.Errorf("commit offsets: %w", err)
	}
	return nil
}

func (c *Consumer) Close() { c.cl.Close() }

func toMessage(rec *kgo.Record) Message {
	msg := Message{Topic: rec.Topic, Key: string(rec.Key), Value: rec.Value}
	if len(rec.Headers) > 0 {
		msg.Headers = make(map[string]string, len(rec.Headers))
		for _, h := range rec.Headers {
			msg.Headers[h.Key] = string(h.Value)
		}
	}
	return msg
}
