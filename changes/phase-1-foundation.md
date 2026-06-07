# Phase 1 — Go-фундамент (общая платформа + CI + базовый compose)

- **Дата:** 2026-06-07
- **Статус:** done
- **Ветка / коммит:** feat/phase-1-foundation (см. `git log`)

## Что сделано
- Создан Go-модуль `services/` (`github.com/honnek/ranked-choice-shop/services`, go 1.21).
- Общий слой `internal/platform/`:
  - `config` — загрузка из env (caarlos0/env), единая структура для всех сервисов;
  - `log` — slog (text в dev, JSON в проде);
  - `postgres` — pgx-пул с ping при старте;
  - `redis` — go-redis клиент с ping;
  - `otel` — инициализация трейсинга через OTLP/HTTP, no-op при пустом endpoint;
  - `httpserver` / `grpcserver` — серверы с остановкой по ctx (graceful);
  - `health` — liveness `/healthz` + readiness `/readyz` с регистрируемыми проверками.
- `cmd/probe` — build-check сервис: поднимает платформу, проверяет redis (и postgres, если задан DSN),
  отдаёт health; имеет режим `-healthcheck` для контейнерного healthcheck в distroless.
- Инфраструктура: multi-stage `Dockerfile` (distroless), `Makefile`, `.golangci.yml`, `.dockerignore`,
  заготовки `buf.yaml`/`buf.gen.yaml`/`sqlc.yaml` для следующих фаз.
- CI: `.github/workflows/go.yml` (build + vet + test, отдельный job golangci-lint).
- `docker-compose.yml`: добавлены `redis` (+ volume `redis_data`) и `probe`; удалён устаревший
  `version:` (в т.ч. из `docker-compose.override.yml`).
- Unit-тесты: `config` (дефолты/переопределения), `health` (live/ready, healthy/down).

## Файлы (добавлено/изменено)
- Добавлено: `services/go.mod`, `services/go.sum`, `services/Makefile`, `services/Dockerfile`,
  `services/.dockerignore`, `services/.golangci.yml`, `services/buf.yaml`, `services/buf.gen.yaml`,
  `services/sqlc.yaml`
- Добавлено: `services/internal/platform/{config,log,postgres,redis,otel,httpserver,grpcserver,health}/*.go`
  (+ тесты `config_test.go`, `health_test.go`)
- Добавлено: `services/cmd/probe/main.go`, `.github/workflows/go.yml`
- Изменено: `docker-compose.yml`, `docker-compose.override.yml`

## Как проверить
```bash
cd services
go build ./... && go vet ./... && go test ./...   # всё зелёное
test -z "$(gofmt -l .)"                            # форматирование чистое
cd ..
docker compose up -d redis                         # redis healthy
docker compose build probe                         # образ собирается
# probe вживую против redis:
REDIS_ADDR=localhost:6379 HTTP_ADDR=:8091 POSTGRES_DSN="" go -C services run ./cmd/probe &
curl -s -o /dev/null -w '%{http_code}\n' localhost:8091/healthz   # 200
curl -s localhost:8091/readyz                                     # {"redis":"ok"}
```

## Результат проверки
- `go build` / `go vet` / `go test` — OK (тесты `config`, `health` проходят).
- `gofmt -l` — пусто (чисто).
- `docker compose build probe` — образ собран (distroless).
- `docker compose up -d redis` — контейнер healthy, `redis-cli ping` → PONG.
- probe локально: `/healthz` → 200, `/readyz` → `{"redis":"ok"}`, SIGTERM → graceful shutdown.
- `golangci-lint` локально не установлен → `make lint` падает на fallback (go vet + gofmt, проходит);
  полный golangci-lint гоняется в CI (`go.yml`, job `lint`). CI на GitHub ещё не прогонялся (нет пуша).

## Remaining
- Нет. Фаза закрыта. Следующая — Фаза 2 (catalog-service).
