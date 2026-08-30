package transport

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderv1 "github.com/honnek/lumewear-shop/services/api/gen/order/v1"
	grpcadapter "github.com/honnek/lumewear-shop/services/internal/order/adapter/grpc"
	"github.com/honnek/lumewear-shop/services/internal/platform/gateway"
)

func NewGateway(ctx context.Context, grpcAddr string) (http.Handler, error) {
	gw := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(headerMatcher))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := orderv1.RegisterOrderServiceHandlerFromEndpoint(ctx, gw, gateway.DialEndpoint(grpcAddr), opts); err != nil {
		return nil, fmt.Errorf("register gateway: %w", err)
	}
	return gw, nil
}

// Idempotency-Key дефолтный матчер не пропускает (заголовок нестандартный), а класть
// ключ в тело REST-клиенту неудобно — пробрасываем его в метаданные как есть.
func headerMatcher(key string) (string, bool) {
	if strings.EqualFold(key, grpcadapter.IdempotencyKeyHeader) {
		return grpcadapter.IdempotencyKeyHeader, true
	}
	return runtime.DefaultHeaderMatcher(key)
}
