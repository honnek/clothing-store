package transport

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	catalogv1 "github.com/honnek/ranked-choice-shop/services/api/gen/catalog/v1"
)

//go:embed openapi/catalog.swagger.json
var openapiSpec []byte

// NewGateway поднимает grpc-gateway, который проксирует REST на gRPC сервиса.
// Соединение ленивое — gRPC-сервер в том же процессе успеет подняться к первому запросу.
func NewGateway(ctx context.Context, grpcAddr string) (http.Handler, error) {
	gw := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := catalogv1.RegisterCatalogServiceHandlerFromEndpoint(ctx, gw, dialEndpoint(grpcAddr), opts); err != nil {
		return nil, fmt.Errorf("register gateway: %w", err)
	}
	return gw, nil
}

// SwaggerHandler отдаёт OpenAPI-спеку и Swagger UI поверх неё. Монтируется под /swagger.
func SwaggerHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(openapiSpec)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(swaggerUIPage)
	})
	return mux
}

// dialEndpoint превращает адрес прослушивания (":9090") в адрес для дозвона gateway.
func dialEndpoint(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "localhost" + addr
	}
	return addr
}

var swaggerUIPage = []byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Catalog API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
  <div id="swagger"></div>
  <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
  <script>SwaggerUIBundle({url: "/swagger/openapi.json", dom_id: "#swagger"});</script>
</body>
</html>`)
