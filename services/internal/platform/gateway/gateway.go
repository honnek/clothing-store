package gateway

import (
	"fmt"
	"net/http"
	"strings"
)

// DialEndpoint превращает адрес прослушивания (":9090") в адрес для дозвона gateway
// до gRPC-сервера в том же процессе.
func DialEndpoint(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "localhost" + addr
	}
	return addr
}

// SwaggerHandler отдаёт OpenAPI-спеку (`/openapi.json`) и Swagger UI поверх неё.
// Монтируется под /swagger; spec вшит в бинарь вызывающей стороной.
func SwaggerHandler(spec []byte, title string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(spec)
	})

	page := []byte(fmt.Sprintf(swaggerPage, title))
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page)
	})
	return mux
}

const swaggerPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>%s</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
  <div id="swagger"></div>
  <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
  <script>SwaggerUIBundle({url: "/swagger/openapi.json", dom_id: "#swagger"});</script>
</body>
</html>`
