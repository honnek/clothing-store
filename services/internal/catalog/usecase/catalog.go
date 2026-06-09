package usecase

import (
	"context"

	"github.com/honnek/ranked-choice-shop/services/internal/catalog/domain"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// Repository — то, что usecase ожидает от слоя хранения. Интерфейс объявлен здесь,
// у потребителя: и postgres-репозиторий, и кеш-декоратор реализуют именно его.
type Repository interface {
	ListProducts(ctx context.Context, f domain.ProductFilter, p domain.Page) (domain.ProductList, error)
	GetProduct(ctx context.Context, uuid string) (domain.Product, error)
	ListCategories(ctx context.Context) ([]domain.Category, error)
}

type Catalog struct {
	repo Repository
}

func New(repo Repository) *Catalog {
	return &Catalog{repo: repo}
}

func (c *Catalog) ListProducts(ctx context.Context, f domain.ProductFilter, p domain.Page) (domain.ProductList, error) {
	// Чиним кривую пагинацию из запроса, чтобы не уехать в полный скан или минус.
	if p.Limit <= 0 || p.Limit > maxLimit {
		p.Limit = defaultLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return c.repo.ListProducts(ctx, f, p)
}

func (c *Catalog) GetProduct(ctx context.Context, uuid string) (domain.Product, error) {
	return c.repo.GetProduct(ctx, uuid)
}

func (c *Catalog) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return c.repo.ListCategories(ctx)
}
