package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/honnek/lumewear-shop/services/internal/catalog/domain"
)

type countingRepo struct {
	getCalls int
	product  domain.Product
}

func (r *countingRepo) ListProducts(context.Context, domain.ProductFilter, domain.Page) (domain.ProductList, error) {
	return domain.ProductList{}, nil
}
func (r *countingRepo) GetProduct(context.Context, string) (domain.Product, error) {
	r.getCalls++
	return r.product, nil
}
func (r *countingRepo) ListCategories(context.Context) ([]domain.Category, error) { return nil, nil }

func TestGetProductReadThrough(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})

	inner := &countingRepo{product: domain.Product{UUID: "u-1", Title: "Hat", Price: "9.99"}}
	c := New(inner, rdb, time.Minute)

	// Первый вызов — промах, идём в репозиторий и кладём в кеш.
	got, err := c.GetProduct(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("first GetProduct: %v", err)
	}
	if got.Title != "Hat" {
		t.Fatalf("title = %q, want Hat", got.Title)
	}

	// Второй вызов — попадание, репозиторий не дёргаем.
	got, err = c.GetProduct(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("second GetProduct: %v", err)
	}
	if got.Price != "9.99" {
		t.Errorf("price = %q, want 9.99", got.Price)
	}
	if inner.getCalls != 1 {
		t.Errorf("inner GetProduct calls = %d, want 1 (second served from cache)", inner.getCalls)
	}
}
