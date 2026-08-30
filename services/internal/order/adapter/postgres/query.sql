-- name: OrderIDByIdempotencyKey :one
SELECT order_id FROM checkout_idempotency WHERE key = sqlc.arg('key');

-- name: LockProductsForCheckout :many
-- ORDER BY id до FOR UPDATE: LockRows стоит над сортировкой, поэтому строки
-- блокируются в одном и том же порядке во всех параллельных checkout-ах — без этого
-- две корзины с пересекающимся составом ловят дедлок.
SELECT
    id,
    uuid::text  AS uuid,
    title,
    price::text AS price,
    stock
FROM product
WHERE uuid = ANY(sqlc.arg('uuids')::uuid[])
  AND is_deleted = false
ORDER BY id
FOR UPDATE;

-- name: ReserveStock :exec
UPDATE product SET stock = stock - sqlc.arg('qty') WHERE id = sqlc.arg('id');

-- name: InsertOrder :one
-- id берём из doctrine-последовательности: таблицу делим с PHP, свои id заводить нельзя.
INSERT INTO "order" (id, owner_id, created_at, status, total_price, updated_at, is_deleted)
VALUES (
    nextval('"order_id_seq"'),
    sqlc.narg('owner_id'),
    now(),
    sqlc.arg('status'),
    sqlc.arg('total_price'),
    now(),
    false
)
RETURNING id, created_at, updated_at;

-- name: InsertOrderProduct :exec
INSERT INTO order_product (id, app_order_id, product_id, quantity, price_per_one)
VALUES (
    nextval('order_product_id_seq'),
    sqlc.arg('app_order_id'),
    sqlc.arg('product_id'),
    sqlc.arg('quantity'),
    sqlc.arg('price_per_one')
);

-- name: InsertIdempotencyKey :exec
INSERT INTO checkout_idempotency (key, session_id, order_id)
VALUES (sqlc.arg('key'), sqlc.arg('session_id'), sqlc.arg('order_id'));

-- name: InsertOutboxEvent :exec
INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
VALUES (
    sqlc.arg('aggregate_type'),
    sqlc.arg('aggregate_id'),
    sqlc.arg('event_type'),
    sqlc.arg('payload')
);

-- name: GetOrder :one
SELECT
    id,
    owner_id,
    status,
    coalesce(total_price, 0)::numeric(10, 2)::text AS total,
    created_at,
    updated_at
FROM "order"
WHERE id = sqlc.arg('id') AND is_deleted = false;

-- name: ListOrderItems :many
SELECT
    op.product_id,
    p.uuid::text            AS uuid,
    p.title,
    op.price_per_one::text  AS price_per_one,
    op.quantity
FROM order_product op
JOIN product p ON p.id = op.product_id
WHERE op.app_order_id = sqlc.arg('app_order_id')
ORDER BY op.id;

-- name: ListOrders :many
SELECT
    id,
    owner_id,
    status,
    coalesce(total_price, 0)::numeric(10, 2)::text AS total,
    created_at,
    updated_at
FROM "order"
WHERE is_deleted = false
  AND (sqlc.narg('owner_id')::int IS NULL OR owner_id = sqlc.narg('owner_id'))
  AND (sqlc.narg('status')::int   IS NULL OR status   = sqlc.narg('status'))
ORDER BY id DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountOrders :one
SELECT count(*)
FROM "order"
WHERE is_deleted = false
  AND (sqlc.narg('owner_id')::int IS NULL OR owner_id = sqlc.narg('owner_id'))
  AND (sqlc.narg('status')::int   IS NULL OR status   = sqlc.narg('status'));

-- name: LockOrderStatus :one
SELECT status FROM "order" WHERE id = sqlc.arg('id') AND is_deleted = false FOR UPDATE;

-- name: UpdateOrderStatus :exec
UPDATE "order" SET status = sqlc.arg('status'), updated_at = now() WHERE id = sqlc.arg('id');
