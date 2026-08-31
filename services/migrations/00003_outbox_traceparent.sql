-- +goose Up

-- Контекст трейса checkout-а, чтобы relay опубликовал событие в том же трейсе, а не
-- завёл оторванный корень: между вставкой и публикацией лежит коммит, и in-process
-- контекст туда не доезжает. Формат — W3C traceparent, его же кладём в заголовок Kafka.
ALTER TABLE outbox ADD COLUMN IF NOT EXISTS traceparent TEXT;

-- +goose Down
ALTER TABLE outbox DROP COLUMN IF EXISTS traceparent;
