package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	catalogv1 "github.com/honnek/lumewear-shop/services/api/gen/catalog/v1"
	"github.com/honnek/lumewear-shop/services/internal/platform/gateway"
)

// NewGateway поднимает grpc-gateway, который проксирует REST на gRPC сервиса.
// Соединение ленивое — gRPC-сервер в том же процессе успеет подняться к первому запросу.
func NewGateway(ctx context.Context, grpcAddr string) (http.Handler, error) {
	gw := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := catalogv1.RegisterCatalogServiceHandlerFromEndpoint(ctx, gw, gateway.DialEndpoint(grpcAddr), opts); err != nil {
		return nil, fmt.Errorf("register gateway: %w", err)
	}
	return gw, nil
}
