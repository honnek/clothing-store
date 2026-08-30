// Package migrations содержит goose-миграции того, что в общей с PHP базе завёл Go:
// остатки товара, ключи идемпотентности checkout и outbox. Таблицы product, "order"
// и order_product по-прежнему принадлежат Doctrine-миграциям монолита.
package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var FS embed.FS

// Up накатывает миграции на базу по dsn. Отдельное database/sql-подключение вместо
// общего pgx-пула: goose умеет только с ним, и живёт оно ровно на время наката.
func Up(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
