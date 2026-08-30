package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config описывает всё, что сервис читает из окружения. Все Go-сервисы репозитория
// грузят одну и ту же структуру — деплой остаётся единообразным, а лишние поля
// сервис просто игнорирует.
type Config struct {
	ServiceName string `env:"SERVICE_NAME" envDefault:"service"`
	Env         string `env:"APP_ENV" envDefault:"dev"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	HTTPAddr string `env:"HTTP_ADDR" envDefault:":8080"`
	GRPCAddr string `env:"GRPC_ADDR" envDefault:":9090"`

	// PostgresDSN опционален до фазы order-service: схема общая с PHP-монолитом,
	// поэтому ранним сервисам БД не нужна.
	PostgresDSN string `env:"POSTGRES_DSN"`

	RedisAddr string `env:"REDIS_ADDR" envDefault:"redis:6379"`
	RedisDB   int    `env:"REDIS_DB" envDefault:"0"`

	// CatalogGRPCAddr — адрес catalog-service для межсервисных gRPC-вызовов (cart/order).
	CatalogGRPCAddr string `env:"CATALOG_GRPC_ADDR" envDefault:"catalog-service:9090"`
	// CartGRPCAddr — адрес cart-service: из него order забирает состав заказа.
	CartGRPCAddr string `env:"CART_GRPC_ADDR" envDefault:"cart-service:9090"`

	// RunMigrations: goose-миграции Go-части накатывает order-service, он ими и владеет.
	RunMigrations bool `env:"RUN_MIGRATIONS" envDefault:"true"`

	// OTLPEndpoint: пустое значение полностью выключает трейсинг (удобно в dev).
	OTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`

	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

func Load() (Config, error) {
	return env.ParseAs[Config]()
}
