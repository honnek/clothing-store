// Command probe — служебный сервис фазы 1: поднимает весь платформенный слой
// (config, логирование, redis, опционально postgres, трейсинг, graceful shutdown)
// и отдаёт health-эндпоинты. Доказывает, что сборка и рантайм-обвязка работают
// сквозняком ещё до появления реальных сервисов. В следующих фазах заменяется.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/honnek/ranked-choice-shop/services/internal/platform/config"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/health"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/httpserver"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/log"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/otel"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/postgres"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/redis"
)

func main() {
	// Healthcheck контейнера: в distroless нет шелла, поэтому бинарь пингует сам себя.
	hcAddr := flag.String("healthcheck", "", "GET http://<addr>/healthz and exit 0/1")
	flag.Parse()
	if *hcAddr != "" {
		os.Exit(selfCheck(*hcAddr))
	}

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "probe:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
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

	hc := health.New()

	rdb, err := redis.New(ctx, cfg.RedisAddr, cfg.RedisDB)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer func() { _ = rdb.Close() }()
	hc.Register("redis", func(ctx context.Context) error { return rdb.Ping(ctx).Err() })

	// Postgres в этой фазе опционален — он нужен только order-service (фаза 4).
	if cfg.PostgresDSN != "" {
		pool, err := postgres.New(ctx, cfg.PostgresDSN)
		if err != nil {
			return fmt.Errorf("connect postgres: %w", err)
		}
		defer pool.Close()
		hc.Register("postgres", pool.Ping)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", hc.Live)
	mux.HandleFunc("/readyz", hc.Ready)

	srv := httpserver.New(cfg.HTTPAddr, mux)
	logger.Info("probe started", "service", cfg.ServiceName, "http", cfg.HTTPAddr)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return srv.Run(gctx, cfg.ShutdownTimeout) })

	if err := g.Wait(); err != nil {
		return err
	}
	logger.Info("probe stopped")
	return nil
}

func selfCheck(addr string) int {
	c := http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get("http://" + addr + "/healthz")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
