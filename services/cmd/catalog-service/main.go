package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"

	catalogv1 "github.com/honnek/ranked-choice-shop/services/api/gen/catalog/v1"
	cacheadapter "github.com/honnek/ranked-choice-shop/services/internal/catalog/adapter/cache"
	grpcadapter "github.com/honnek/ranked-choice-shop/services/internal/catalog/adapter/grpc"
	pgadapter "github.com/honnek/ranked-choice-shop/services/internal/catalog/adapter/postgres"
	"github.com/honnek/ranked-choice-shop/services/internal/catalog/transport"
	"github.com/honnek/ranked-choice-shop/services/internal/catalog/usecase"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/config"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/grpcserver"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/health"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/httpserver"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/log"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/otel"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/postgres"
	"github.com/honnek/ranked-choice-shop/services/internal/platform/redis"
)

const cacheTTL = time.Minute

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "catalog-service:", err)
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

	pool, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	rdb, err := redis.New(ctx, cfg.RedisAddr, cfg.RedisDB)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer func() { _ = rdb.Close() }()

	// Кеш оборачивает postgres; usecase знает только про интерфейс Repository.
	repo := cacheadapter.New(pgadapter.New(pool), rdb, cacheTTL)
	svc := usecase.New(repo)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	grpcSrv := grpcserver.New(lis)
	catalogv1.RegisterCatalogServiceServer(grpcSrv.Grpc(), grpcadapter.NewServer(svc))
	reflection.Register(grpcSrv.Grpc())

	hc := health.New()
	hc.Register("postgres", pool.Ping)
	hc.Register("redis", func(ctx context.Context) error { return rdb.Ping(ctx).Err() })

	gw, err := transport.NewGateway(ctx, cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("gateway: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", hc.Live)
	mux.HandleFunc("/readyz", hc.Ready)
	mux.Handle("/v1/", gw)
	mux.Handle("/swagger/", http.StripPrefix("/swagger", transport.SwaggerHandler()))
	httpSrv := httpserver.New(cfg.HTTPAddr, mux)

	logger.Info("catalog-service started", "grpc", cfg.GRPCAddr, "http", cfg.HTTPAddr)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return grpcSrv.Run(gctx) })
	g.Go(func() error { return httpSrv.Run(gctx, cfg.ShutdownTimeout) })

	if err := g.Wait(); err != nil {
		return err
	}
	logger.Info("catalog-service stopped")
	return nil
}
