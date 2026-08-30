-- Срез таблиц монолита, которые нужны заказу. Владелец — Doctrine; здесь снимок
-- для типов sqlc и для инициализации тестовой базы. Всё, что заводит Go
-- (product.stock, checkout_idempotency, outbox), лежит в services/migrations.

CREATE SEQUENCE "order_id_seq" INCREMENT BY 1 MINVALUE 1 START 1;
CREATE SEQUENCE order_product_id_seq INCREMENT BY 1 MINVALUE 1 START 1;

CREATE TABLE category (
    id    INT          NOT NULL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    slug  VARCHAR(120) NOT NULL
);

CREATE TABLE product (
    id           INT           NOT NULL PRIMARY KEY,
    uuid         UUID,
    title        VARCHAR(255)  NOT NULL,
    price        NUMERIC(6, 2) NOT NULL,
    quality      INT           NOT NULL,
    created_at   TIMESTAMP(0)  WITHOUT TIME ZONE NOT NULL,
    description  TEXT,
    is_published BOOLEAN       NOT NULL,
    is_deleted   BOOLEAN       NOT NULL,
    slug         VARCHAR(128),
    category_id  INT REFERENCES category (id)
);

CREATE TABLE "order" (
    id          INT          NOT NULL PRIMARY KEY,
    owner_id    INT,
    created_at  TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    status      INT          NOT NULL,
    total_price DOUBLE PRECISION,
    updated_at  TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    is_deleted  BOOLEAN      NOT NULL
);

CREATE TABLE order_product (
    id            INT           NOT NULL PRIMARY KEY,
    app_order_id  INT           NOT NULL REFERENCES "order" (id),
    product_id    INT           NOT NULL REFERENCES product (id),
    quantity      INT           NOT NULL,
    price_per_one NUMERIC(6, 2) NOT NULL
);

CREATE INDEX order_product_app_order_idx ON order_product (app_order_id);
