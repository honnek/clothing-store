-- +goose Up

-- Остатки живут в product: заказ и резерв идут одной транзакцией, отдельная таблица
-- складов дала бы лишний join под блокировкой без выигрыша.
ALTER TABLE product ADD COLUMN IF NOT EXISTS stock INT NOT NULL DEFAULT 0;
ALTER TABLE product ADD CONSTRAINT product_stock_non_negative CHECK (stock >= 0);

-- Ключ идемпотентности checkout: PK по key — единственный барьер против двойного
-- оформления при ретрае клиента и при двух параллельных запросах с одним ключом.
CREATE TABLE checkout_idempotency (
    key        TEXT        NOT NULL PRIMARY KEY,
    session_id TEXT        NOT NULL,
    order_id   INT         NOT NULL REFERENCES "order" (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Transactional outbox: событие пишется в той же транзакции, что и заказ,
-- relay публикует его в Kafka уже после коммита (фаза 6).
CREATE TABLE outbox (
    id             BIGSERIAL   NOT NULL PRIMARY KEY,
    aggregate_type TEXT        NOT NULL,
    aggregate_id   TEXT        NOT NULL,
    event_type     TEXT        NOT NULL,
    payload        JSONB       NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ
);

-- Частичный индекс: relay читает только неопубликованное, опубликованное копится и не мешает.
CREATE INDEX outbox_pending_idx ON outbox (id) WHERE published_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS outbox_pending_idx;
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS checkout_idempotency;
ALTER TABLE product DROP CONSTRAINT IF EXISTS product_stock_non_negative;
ALTER TABLE product DROP COLUMN IF EXISTS stock;
