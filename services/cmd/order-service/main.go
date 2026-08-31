package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	orderv1 "github.com/honnek/lumewear-shop/services/api/gen/order/v1"
	"github.com/honnek/lumewear-shop/services/api/openapi"
	"github.com/honnek/lumewear-shop/services/internal/order/adapter/cartclient"
	grpcadapter "github.com/honnek/lumewear-shop/services/internal/order/adapter/grpc"
	"github.com/honnek/lumewear-shop/services/internal/order/adapter/outbox"
	pgadapter "github.com/honnek/lumewear-shop/services/internal/order/adapter/postgres"
	"github.com/honnek/lumewear-shop/services/internal/order/transport"
	"github.com/honnek/lumewear-shop/services/internal/order/usecase"
	"github.com/honnek/lumewear-shop/services/internal/platform/config"
	"github.com/honnek/lumewear-shop/services/internal/platform/gateway"
	"github.com/honnek/lumewear-shop/services/internal/platform/grpcserver"
	"github.com/honnek/lumewear-shop/services/internal/platform/health"
	"github.com/honnek/lumewear-shop/services/internal/platform/httpserver"
	"github.com/honnek/lumewear-shop/services/internal/platform/kafka"
	"github.com/honnek/lumewear-shop/services/internal/platform/log"
	"github.com/honnek/lumewear-shop/services/internal/platform/metrics"
	"github.com/honnek/lumewear-shop/services/internal/platform/otel"
	"github.com/honnek/lumewear-shop/services/internal/platform/postgres"
	"github.com/honnek/lumewear-shop/services/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "order-service:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.PostgresDSN == "" {
		return fmt.Errorf("POSTGRES_DSN is required")
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

	if cfg.RunMigrations {
		if err := migrations.Up(cfg.PostgresDSN); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
		logger.Info("migrations applied")
	}

	pool, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	cartConn, err := grpc.NewClient(cfg.CartGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("dial cart: %w", err)
	}
	defer func() { _ = cartConn.Close() }()

	repo := pgadapter.New(pool)
	svc := usecase.New(repo, cartclient.New(cartConn), logger)

	// Relay живёт в том же процессе, что и checkout: отдельный бинарь ничего не
	// упрощает, а SKIP LOCKED в запросе разводит реплики без координации.
	var relay *outbox.Relay
	if len(cfg.KafkaBrokers) > 0 {
		producer, err := kafka.NewProducer(cfg.KafkaBrokers)
		if err != nil {
			return err
		}
		defer producer.Close()

		relay = outbox.NewRelay(repo, producer, logger, cfg.OrderTopic, cfg.OutboxBatch, cfg.OutboxInterval)
	} else {
		logger.Warn("KAFKA_BROKERS is empty, outbox relay disabled")
	}

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	grpcSrv := grpcserver.New(lis,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()),
	)
	orderv1.RegisterOrderServiceServer(grpcSrv.Grpc(), grpcadapter.NewServer(svc))
	reflection.Register(grpcSrv.Grpc())

	gw, err := transport.NewGateway(ctx, cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("gateway: %w", err)
	}

	hc := health.New()
	hc.Register("postgres", pool.Ping)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", hc.Live)
	mux.HandleFunc("/readyz", hc.Ready)
	mux.Handle("/metrics", metrics.Handler())
	mux.Handle("/v1/", gw)
	mux.Handle("/swagger/", http.StripPrefix("/swagger", gateway.SwaggerHandler(openapi.Order, "Order API")))
	httpSrv := httpserver.New(cfg.HTTPAddr, mux)

	logger.Info("order-service started", "grpc", cfg.GRPCAddr, "http", cfg.HTTPAddr, "cart", cfg.CartGRPCAddr)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return grpcSrv.Run(gctx) })
	g.Go(func() error { return httpSrv.Run(gctx, cfg.ShutdownTimeout) })
	if relay != nil {
		g.Go(func() error { return relay.Run(gctx) })
	}

	if err := g.Wait(); err != nil {
		return err
	}
	logger.Info("order-service stopped")
	return nil
}
