-- name: ListProducts :many
-- uuid/price приводим к text: pgtype-обёртки наружу не нужны, домен оперирует строками.
SELECT
    id,
    uuid::text  AS uuid,
    title,
    price::text AS price,
    quality,
    description,
    is_published,
    slug,
    category_id
FROM product
WHERE is_deleted = false
  AND (sqlc.narg('category_id')::int  IS NULL OR category_id  = sqlc.narg('category_id'))
  AND (sqlc.narg('published')::bool   IS NULL OR is_published = sqlc.narg('published'))
  AND (sqlc.narg('search')::text      IS NULL OR title ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY id DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountProducts :one
SELECT count(*)
FROM product
WHERE is_deleted = false
  AND (sqlc.narg('category_id')::int  IS NULL OR category_id  = sqlc.narg('category_id'))
  AND (sqlc.narg('published')::bool   IS NULL OR is_published = sqlc.narg('published'))
  AND (sqlc.narg('search')::text      IS NULL OR title ILIKE '%' || sqlc.narg('search') || '%');

-- name: GetProductByUUID :one
SELECT
    id,
    uuid::text  AS uuid,
    title,
    price::text AS price,
    quality,
    description,
    is_published,
    slug,
    category_id
FROM product
WHERE uuid = sqlc.arg('product_uuid') AND is_deleted = false;

-- name: ListCategories :many
SELECT id, title, slug
FROM category
ORDER BY title;
