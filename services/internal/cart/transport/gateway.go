package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cartv1 "github.com/honnek/lumewear-shop/services/api/gen/cart/v1"
	"github.com/honnek/lumewear-shop/services/internal/platform/gateway"
)

func NewGateway(ctx context.Context, grpcAddr string) (http.Handler, error) {
	gw := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := cartv1.RegisterCartServiceHandlerFromEndpoint(ctx, gw, gateway.DialEndpoint(grpcAddr), opts); err != nil {
		return nil, fmt.Errorf("register gateway: %w", err)
	}
	return gw, nil
}
