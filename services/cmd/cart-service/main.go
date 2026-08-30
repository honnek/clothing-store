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

	cartv1 "github.com/honnek/lumewear-shop/services/api/gen/cart/v1"
	"github.com/honnek/lumewear-shop/services/api/openapi"
	"github.com/honnek/lumewear-shop/services/internal/cart/adapter/catalogclient"
	grpcadapter "github.com/honnek/lumewear-shop/services/internal/cart/adapter/grpc"
	redisrepo "github.com/honnek/lumewear-shop/services/internal/cart/adapter/redis"
	"github.com/honnek/lumewear-shop/services/internal/cart/transport"
	"github.com/honnek/lumewear-shop/services/internal/cart/usecase"
	"github.com/honnek/lumewear-shop/services/internal/platform/config"
	"github.com/honnek/lumewear-shop/services/internal/platform/gateway"
	"github.com/honnek/lumewear-shop/services/internal/platform/grpcserver"
	"github.com/honnek/lumewear-shop/services/internal/platform/health"
	"github.com/honnek/lumewear-shop/services/internal/platform/httpserver"
	"github.com/honnek/lumewear-shop/services/internal/platform/log"
	"github.com/honnek/lumewear-shop/services/internal/platform/metrics"
	"github.com/honnek/lumewear-shop/services/internal/platform/otel"
	"github.com/honnek/lumewear-shop/services/internal/platform/redis"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "cart-service:", err)
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

	rdb, err := redis.New(ctx, cfg.RedisAddr, cfg.RedisDB)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer func() { _ = rdb.Close() }()

	catalogConn, err := grpc.NewClient(cfg.CatalogGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("dial catalog: %w", err)
	}
	defer func() { _ = catalogConn.Close() }()

	svc := usecase.New(redisrepo.New(rdb), catalogclient.New(catalogConn))

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	grpcSrv := grpcserver.New(lis,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()),
	)
	cartv1.RegisterCartServiceServer(grpcSrv.Grpc(), grpcadapter.NewServer(svc))
	reflection.Register(grpcSrv.Grpc())

	gw, err := transport.NewGateway(ctx, cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("gateway: %w", err)
	}

	hc := health.New()
	hc.Register("redis", func(ctx context.Context) error { return rdb.Ping(ctx).Err() })

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", hc.Live)
	mux.HandleFunc("/readyz", hc.Ready)
	mux.Handle("/metrics", metrics.Handler())
	mux.Handle("/v1/", gw)
	mux.Handle("/swagger/", http.StripPrefix("/swagger", gateway.SwaggerHandler(openapi.Cart, "Cart API")))
	httpSrv := httpserver.New(cfg.HTTPAddr, mux)

	logger.Info("cart-service started", "grpc", cfg.GRPCAddr, "http", cfg.HTTPAddr, "catalog", cfg.CatalogGRPCAddr)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return grpcSrv.Run(gctx) })
	g.Go(func() error { return httpSrv.Run(gctx, cfg.ShutdownTimeout) })

	if err := g.Wait(); err != nil {
		return err
	}
	logger.Info("cart-service stopped")
	return nil
}
