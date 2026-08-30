# Lumewear Shop

![Go](https://img.shields.io/badge/Go-1.21-00ADD8?logo=go&logoColor=white)
![PHP](https://img.shields.io/badge/PHP-8.1-777BB4?logo=php&logoColor=white)
![Symfony](https://img.shields.io/badge/Symfony-6-000000?logo=symfony&logoColor=white)
![CI](https://github.com/honnek/clothing-store/actions/workflows/go.yml/badge.svg)
![License](https://img.shields.io/badge/License-MIT-green)

> **EN** — An e-commerce app whose customer-facing core is being migrated from a Symfony/Vue monolith to Go microservices (strangler-fig). PHP stays for admin/CMS, auth and rendering; Go owns catalog, cart, checkout and order lifecycle.
>
> **RU** — Интернет-магазин: клиентское ядро переезжает из монолита Symfony/Vue в Go-микросервисы (strangler-fig). PHP остаётся под admin/CMS, авторизацию и рендеринг; Go берёт каталог, корзину, checkout и жизненный цикл заказа.

**English** · [Русский](#русский)

---

## English

### Overview

This started as a working Symfony 6 + Vue 2 + PostgreSQL online store. It is being evolved into a polyglot system: the hot, customer-facing domain is extracted into Go microservices, while the existing PHP monolith keeps the parts where it is still a good fit (admin/CMS, authentication, server-side rendering, i18n).

The migration follows the **strangler-fig** pattern — a realistic, incremental path used in production when moving load off a legacy monolith, rather than a risky big-bang rewrite.

### Why this design

- **Real domain, real trade-offs.** An e-commerce checkout naturally exercises transactions, concurrency, idempotency and event-driven flows — the topics that actually come up in senior interviews.
- **Concrete value, not a like-for-like port.** The original PHP checkout never reserved stock (the "two buyers, last item" race was broken). The Go `order-service` fixes this with atomic stock reservation under concurrency plus idempotent checkout.
- **Shared PostgreSQL.** Go services read the existing tables and own the order tables with their own migrations — exactly how an incremental extraction looks in practice.

### Services

| Service                | Role                                                            |
| ---------------------- | -------------------------------------------------------------- |
| `catalog-service`      | Product/category read API, search & filters, Redis cache        |
| `cart-service`         | Redis-backed cart, calls catalog for price/availability         |
| `order-service`        | Idempotent checkout, stock reservation, order state machine, outbox |
| `notifications-worker` | Consumes `order.created` from Kafka, sends email                |
| PHP Symfony (kept)     | admin/CMS, auth (Google OAuth), front-end rendering, i18n        |

### Tech stack

**Go:** gRPC + grpc-gateway, pgx + sqlc + goose, go-redis, Kafka (Redpanda) via franz-go, OpenTelemetry + Prometheus + slog, testify + testcontainers, golangci-lint + buf.
**PHP:** Symfony 6, Doctrine, API Platform, Twig + Vue 2.
**Infra:** Docker / docker-compose, GitHub Actions.

### Run locally

```bash
docker compose up -d redis
docker compose up -d --build probe        # phase-1 build-check service
curl localhost:8090/healthz               # {"status":"ok"}
curl localhost:8090/readyz                # {"redis":"ok","postgres":"ok"}
```

Go workspace lives in [`services/`](services/):

```bash
cd services
make test     # unit tests
make lint     # golangci-lint (falls back to go vet + gofmt)
```

### Repository layout

```
services/            Go microservices (clean architecture, shared internal/platform)
src/                 PHP Symfony app (admin/CMS, auth, rendering)
docker-compose.yml   full local stack
docs/legacy.md       original project notes
```

---

## Русский

### Обзор

Изначально — рабочий интернет-магазин на Symfony 6 + Vue 2 + PostgreSQL. Превращается в полиглот-систему: «горячее» клиентское ядро выносится в Go-микросервисы, а монолит на PHP остаётся там, где он уместен (admin/CMS, авторизация, серверный рендеринг, i18n).

Миграция идёт по паттерну **strangler-fig** — реалистичный пошаговый способ снимать нагрузку с легаси-монолита, без рискованного переписывания «всё и сразу».

### Почему так

- **Настоящий домен и настоящие компромиссы.** Оформление заказа само по себе задействует транзакции, конкурентность, идемпотентность и событийную обработку — то, что реально спрашивают на собеседованиях senior-уровня.
- **Конкретная польза, а не калька.** В исходном PHP-чекауте остатки не резервировались (гонка «два покупателя на последний товар» была сломана). Go `order-service` это чинит: атомарный резерв остатков под конкурентностью + идемпотентный checkout.
- **Общая PostgreSQL.** Go-сервисы читают существующие таблицы и владеют таблицами заказов со своими миграциями — ровно так выглядит постепенный вынос на практике.

### Сервисы

| Сервис                 | Роль                                                                  |
| ---------------------- | -------------------------------------------------------------------- |
| `catalog-service`      | Read-API товаров/категорий, поиск и фильтры, Redis-кеш               |
| `cart-service`         | Корзина в Redis, ходит в catalog за ценой/наличием                   |
| `order-service`        | Идемпотентный checkout, резерв остатков, статус-машина заказа, outbox |
| `notifications-worker` | Читает `order.created` из Kafka, шлёт письма                          |
| PHP Symfony (остаётся) | admin/CMS, авторизация (Google OAuth), рендеринг фронта, i18n         |

### Стек

**Go:** gRPC + grpc-gateway, pgx + sqlc + goose, go-redis, Kafka (Redpanda) через franz-go, OpenTelemetry + Prometheus + slog, testify + testcontainers, golangci-lint + buf.
**PHP:** Symfony 6, Doctrine, API Platform, Twig + Vue 2.
**Инфра:** Docker / docker-compose, GitHub Actions.

### Запуск локально

```bash
docker compose up -d redis
docker compose up -d --build probe        # build-check сервис фазы 1
curl localhost:8090/healthz               # {"status":"ok"}
curl localhost:8090/readyz                # {"redis":"ok","postgres":"ok"}
```

Go-воркспейс — в [`services/`](services/):

```bash
cd services
make test     # unit-тесты
make lint     # golangci-lint (откат на go vet + gofmt)
```

### Структура репозитория

```
services/            Go-микросервисы (clean architecture, общий internal/platform)
src/                 PHP-приложение Symfony (admin/CMS, auth, рендеринг)
docker-compose.yml   полный локальный стек
docs/legacy.md       исходные заметки по проекту
```
