package usecase

import (
	"context"
	"testing"

	"github.com/honnek/ranked-choice-shop/services/internal/catalog/domain"
)

type fakeRepo struct {
	gotPage domain.Page
}

func (f *fakeRepo) ListProducts(_ context.Context, _ domain.ProductFilter, p domain.Page) (domain.ProductList, error) {
	f.gotPage = p
	return domain.ProductList{}, nil
}
func (f *fakeRepo) GetProduct(context.Context, string) (domain.Product, error) {
	return domain.Product{}, nil
}
func (f *fakeRepo) ListCategories(context.Context) ([]domain.Category, error) { return nil, nil }

func TestListProductsClampsPagination(t *testing.T) {
	tests := []struct {
		name       string
		in         domain.Page
		wantLimit  int32
		wantOffset int32
	}{
		{"zero limit -> default", domain.Page{Limit: 0, Offset: 0}, defaultLimit, 0},
		{"over max -> default", domain.Page{Limit: 999, Offset: 5}, defaultLimit, 5},
		{"valid stays", domain.Page{Limit: 50, Offset: 10}, 50, 10},
		{"negative offset -> 0", domain.Page{Limit: 10, Offset: -3}, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			c := New(repo)

			if _, err := c.ListProducts(context.Background(), domain.ProductFilter{}, tt.in); err != nil {
				t.Fatalf("ListProducts: %v", err)
			}
			if repo.gotPage.Limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", repo.gotPage.Limit, tt.wantLimit)
			}
			if repo.gotPage.Offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", repo.gotPage.Offset, tt.wantOffset)
			}
		})
	}
}
