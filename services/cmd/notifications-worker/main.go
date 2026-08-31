package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	notifyredis "github.com/honnek/lumewear-shop/services/internal/notifications/adapter/redis"
	"github.com/honnek/lumewear-shop/services/internal/notifications/adapter/smtp"
	"github.com/honnek/lumewear-shop/services/internal/notifications/usecase"
	"github.com/honnek/lumewear-shop/services/internal/platform/config"
	"github.com/honnek/lumewear-shop/services/internal/platform/health"
	"github.com/honnek/lumewear-shop/services/internal/platform/httpserver"
	"github.com/honnek/lumewear-shop/services/internal/platform/kafka"
	"github.com/honnek/lumewear-shop/services/internal/platform/log"
	"github.com/honnek/lumewear-shop/services/internal/platform/metrics"
	"github.com/honnek/lumewear-shop/services/internal/platform/otel"
	"github.com/honnek/lumewear-shop/services/internal/platform/redis"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "notifications-worker:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if len(cfg.KafkaBrokers) == 0 {
		return fmt.Errorf("KAFKA_BROKERS is required")
	}

	logger := log.New(cfg.LogLevel, cfg.Env)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	flushTracing, err := otel.Setup(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
	if err != nil {
		return fmt.Errorf("setup tracing: %w", err)
	}
	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := flushTracing(shutCtx); err != nil {
			logger.Warn("tracing shutdown", "err", err)
		}
	}()

	rdb, err := redis.New(ctx, cfg.RedisAddr, cfg.RedisDB)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer func() { _ = rdb.Close() }()

	consumer, err := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaGroup, logger, cfg.OrderTopic)
	if err != nil {
		return err
	}
	defer consumer.Close()

	notifier := usecase.New(
		smtp.New(cfg.SMTPAddr, cfg.MailFrom),
		notifyredis.NewDedup(rdb, cfg.NotifyTTL),
		logger,
		cfg.MailTo,
	)

	hc := health.New()
	hc.Register("redis", func(ctx context.Context) error { return rdb.Ping(ctx).Err() })

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", hc.Live)
	mux.HandleFunc("/readyz", hc.Ready)
	mux.Handle("/metrics", metrics.Handler())
	httpSrv := httpserver.New(cfg.HTTPAddr, mux)

	logger.Info("notifications-worker started",
		"topic", cfg.OrderTopic, "group", cfg.KafkaGroup, "smtp", cfg.SMTPAddr)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return consumer.Run(gctx, notifier.Handle) })
	g.Go(func() error { return httpSrv.Run(gctx, cfg.ShutdownTimeout) })

	if err := g.Wait(); err != nil {
		return err
	}
	logger.Info("notifications-worker stopped")
	return nil
}
