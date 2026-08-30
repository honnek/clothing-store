GO ?= go

POSTGRES_DSN ?= postgres://postgres:postgres@localhost:5432/lumewear_2?sslmode=disable

.PHONY: build vet test test-integration lint fmt tidy migrate

build:
	$(GO) build ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

# Тесты на testcontainers (docker), под build-тегом integration.
test-integration:
	$(GO) test -tags=integration ./...

fmt:
	gofmt -w .

# Предпочитаем golangci-lint; если его нет — откатываемся на vet + проверку gofmt,
# чтобы `make lint` работал и на голой машине.
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found; running go vet + gofmt check"; \
		$(GO) vet ./...; \
		test -z "$$(gofmt -l .)" || { echo "gofmt needed in:"; gofmt -l .; exit 1; }; \
	fi

tidy:
	$(GO) mod tidy

# Накатывает goose-миграции заказа (остатки, идемпотентность, outbox) на общую с PHP базу.
migrate:
	goose -dir migrations postgres "$(POSTGRES_DSN)" up
