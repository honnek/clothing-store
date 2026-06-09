//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/honnek/ranked-choice-shop/services/internal/catalog/domain"
)

// setupRepo поднимает postgres в контейнере, накатывает схему и сид, отдаёт репозиторий.
func setupRepo(t *testing.T) *Repository {
	t.Helper()
	ctx := context.Background()

	pg, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpostgres.WithDatabase("catalog"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithInitScripts("schema.sql", "testdata/seed.sql"),
		// postgres логирует «ready» сначала для временного сервера под init-скрипты,
		// и только второй раз — когда реально принимает подключения. Ждём второго.
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	return New(pool)
}

func TestListProducts(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()
	page := domain.Page{Limit: 10, Offset: 0}

	t.Run("hides deleted, newest first", func(t *testing.T) {
		list, err := repo.ListProducts(ctx, domain.ProductFilter{}, page)
		if err != nil {
			t.Fatal(err)
		}
		if list.Total != 3 {
			t.Fatalf("total = %d, want 3 (deleted excluded)", list.Total)
		}
		if len(list.Items) != 3 || list.Items[0].ID != 3 {
			t.Fatalf("expected newest-first [3,2,1], got %+v", ids(list.Items))
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		cat := int32(1)
		list, err := repo.ListProducts(ctx, domain.ProductFilter{CategoryID: &cat}, page)
		if err != nil {
			t.Fatal(err)
		}
		if list.Total != 2 {
			t.Fatalf("total = %d, want 2", list.Total)
		}
	})

	t.Run("filter by published", func(t *testing.T) {
		pub := false
		list, err := repo.ListProducts(ctx, domain.ProductFilter{Published: &pub}, page)
		if err != nil {
			t.Fatal(err)
		}
		if list.Total != 1 || list.Items[0].ID != 3 {
			t.Fatalf("want only unpublished id=3, got %+v", ids(list.Items))
		}
	})

	t.Run("search is case-insensitive", func(t *testing.T) {
		q := "hat"
		list, err := repo.ListProducts(ctx, domain.ProductFilter{Search: &q}, page)
		if err != nil {
			t.Fatal(err)
		}
		if list.Total != 2 {
			t.Fatalf("total = %d, want 2 (both hats)", list.Total)
		}
	})

	t.Run("pagination limit", func(t *testing.T) {
		list, err := repo.ListProducts(ctx, domain.ProductFilter{}, domain.Page{Limit: 2, Offset: 0})
		if err != nil {
			t.Fatal(err)
		}
		if len(list.Items) != 2 {
			t.Fatalf("items = %d, want 2", len(list.Items))
		}
		if list.Total != 3 {
			t.Fatalf("total = %d, want 3 (full count ignores limit)", list.Total)
		}
	})
}

func TestGetProduct(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		p, err := repo.GetProduct(ctx, "11111111-1111-1111-1111-111111111111")
		if err != nil {
			t.Fatal(err)
		}
		if p.Title != "Red Hat" || p.Price != "9.99" {
			t.Fatalf("got %+v", p)
		}
	})

	t.Run("deleted is not found", func(t *testing.T) {
		_, err := repo.GetProduct(ctx, "44444444-4444-4444-4444-444444444444")
		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Fatalf("err = %v, want ErrProductNotFound", err)
		}
	})

	t.Run("malformed uuid is not found", func(t *testing.T) {
		_, err := repo.GetProduct(ctx, "not-a-uuid")
		if !errors.Is(err, domain.ErrProductNotFound) {
			t.Fatalf("err = %v, want ErrProductNotFound", err)
		}
	})
}

func TestListCategories(t *testing.T) {
	repo := setupRepo(t)

	cats, err := repo.ListCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 2 || cats[0].Title != "Hats" {
		t.Fatalf("want [Hats, Shoes] ordered, got %+v", cats)
	}
}

func ids(ps []domain.Product) []int32 {
	out := make([]int32, len(ps))
	for i, p := range ps {
		out[i] = p.ID
	}
	return out
}
