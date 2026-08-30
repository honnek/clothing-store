// Package metrics отдаёт Prometheus-эндпоинт и инструментовку gRPC-сервера.
package metrics

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_server_requests_total",
		Help: "Обработанные gRPC-вызовы в разрезе метода и кода ответа.",
	}, []string{"service", "method", "code"})

	duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grpc_server_request_duration_seconds",
		Help:    "Длительность обработки gRPC-вызова.",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method"})
)

func init() {
	prometheus.MustRegister(requests, duration)
}

// Handler — /metrics со стандартными go/process-коллекторами плюс нашими gRPC.
func Handler() http.Handler {
	return promhttp.Handler()
}

// UnaryServerInterceptor считает вызовы и их длительность. Стриминга у сервисов нет,
// поэтому отдельного stream-перехватчика не заводим.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		svc, method := split(info.FullMethod)

		start := time.Now()
		resp, err := handler(ctx, req)
		duration.WithLabelValues(svc, method).Observe(time.Since(start).Seconds())
		requests.WithLabelValues(svc, method, status.Code(err).String()).Inc()

		return resp, err
	}
}

// FullMethod приходит как "/order.v1.OrderService/Checkout".
func split(fullMethod string) (service, method string) {
	trimmed := strings.TrimPrefix(fullMethod, "/")
	if i := strings.LastIndex(trimmed, "/"); i >= 0 {
		return trimmed[:i], trimmed[i+1:]
	}
	return trimmed, ""
}
