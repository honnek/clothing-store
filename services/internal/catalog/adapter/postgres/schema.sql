-- Срез существующих таблиц магазина, которыми владеет PHP-монолит.
-- Нужен только для генерации типов sqlc; миграциями отсюда не управляем.

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
