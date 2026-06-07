# Changes — журнал реализации по фазам

Прогресс миграции покупательского ядра на Go (strangler-fig поверх Symfony-монолита).
План: `~/.claude/plans/harmonic-herding-kahn.md`. Каждую фазу выполняет skill `/next-phase`,
который кладёт сюда детальный отчёт `phase-N-<slug>.md` и отмечает чекбокс ниже.

Легенда: `[ ]` не начато · `[~]` частично (см. Remaining в файле фазы) · `[x]` готово.

## Прогресс

- [x] **Фаза 1** — Go-фундамент (общая платформа + CI + базовый compose) → [phase-1-foundation.md](phase-1-foundation.md)
- [ ] **Фаза 2** — catalog-service (read + search + Redis-кеш)
- [ ] **Фаза 3** — cart-service (Redis-backed корзина)
- [ ] **Фаза 4** — order-service: БД, домен, idempotent checkout (ядро)
- [ ] **Фаза 5** — order-service: gRPC + REST-gateway + Swagger
- [ ] **Фаза 6** — события (Kafka + outbox + worker) + observability
- [ ] **Фаза 7** — интеграция PHP→Go + README + финал

## Отчёты по фазам

_(появятся здесь по мере выполнения: `phase-1-foundation.md`, `phase-2-catalog.md`, …)_
