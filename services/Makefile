GO ?= go

.PHONY: build vet test lint fmt tidy

build:
	$(GO) build ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

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
