-- +goose Up

-- Первичное наполнение остатков: до этой фазы количества не было вообще,
-- поэтому все живые товары стартуют с одинаковой партии.
UPDATE product SET stock = 10 WHERE stock = 0 AND is_deleted = false;

-- +goose Down
SELECT 1;
