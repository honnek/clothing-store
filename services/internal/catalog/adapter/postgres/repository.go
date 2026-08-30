package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/honnek/lumewear-shop/services/internal/catalog/adapter/postgres/catalogdb"
	"github.com/honnek/lumewear-shop/services/internal/catalog/domain"
)

type Repository struct {
	q *catalogdb.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{q: catalogdb.New(pool)}
}

func (r *Repository) ListProducts(ctx context.Context, f domain.ProductFilter, p domain.Page) (domain.ProductList, error) {
	items, err := r.q.ListProducts(ctx, catalogdb.ListProductsParams{
		CategoryID: f.CategoryID,
		Published:  f.Published,
		Search:     f.Search,
		Lim:        p.Limit,
		Off:        p.Offset,
	})
	if err != nil {
		return domain.ProductList{}, fmt.Errorf("list products: %w", err)
	}

	total, err := r.q.CountProducts(ctx, catalogdb.CountProductsParams{
		CategoryID: f.CategoryID,
		Published:  f.Published,
		Search:     f.Search,
	})
	if err != nil {
		return domain.ProductList{}, fmt.Errorf("count products: %w", err)
	}

	out := make([]domain.Product, 0, len(items))
	for _, it := range items {
		out = append(out, domain.Product{
			ID:          it.ID,
			UUID:        it.Uuid,
			Title:       it.Title,
			Price:       it.Price,
			Quality:     it.Quality,
			Description: deref(it.Description),
			Slug:        deref(it.Slug),
			CategoryID:  it.CategoryID,
			IsPublished: it.IsPublished,
		})
	}
	return domain.ProductList{Items: out, Total: total}, nil
}

func (r *Repository) GetProduct(ctx context.Context, uuid string) (domain.Product, error) {
	var pu pgtype.UUID
	if err := pu.Scan(uuid); err != nil {
		// Кривой uuid в запросе — для клиента это просто «не найдено».
		return domain.Product{}, domain.ErrProductNotFound
	}

	row, err := r.q.GetProductByUUID(ctx, pu)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Product{}, domain.ErrProductNotFound
	}
	if err != nil {
		return domain.Product{}, fmt.Errorf("get product: %w", err)
	}

	return domain.Product{
		ID:          row.ID,
		UUID:        row.Uuid,
		Title:       row.Title,
		Price:       row.Price,
		Quality:     row.Quality,
		Description: deref(row.Description),
		Slug:        deref(row.Slug),
		CategoryID:  row.CategoryID,
		IsPublished: row.IsPublished,
	}, nil
}

func (r *Repository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	out := make([]domain.Category, 0, len(rows))
	for _, c := range rows {
		out = append(out, domain.Category{ID: c.ID, Title: c.Title, Slug: c.Slug})
	}
	return out, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
