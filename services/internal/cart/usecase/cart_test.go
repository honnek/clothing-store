package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/honnek/lumewear-shop/services/internal/cart/domain"
)

type memRepo struct {
	items map[string]int32
}

func newMemRepo() *memRepo { return &memRepo{items: map[string]int32{}} }

func (m *memRepo) Add(_ context.Context, _, uuid string, delta int32) error {
	m.items[uuid] += delta
	return nil
}
func (m *memRepo) SetQty(_ context.Context, _, uuid string, qty int32) error {
	m.items[uuid] = qty
	return nil
}
func (m *memRepo) Remove(_ context.Context, _, uuid string) error {
	delete(m.items, uuid)
	return nil
}
func (m *memRepo) Items(_ context.Context, _ string) (map[string]int32, error) {
	return m.items, nil
}

type fakeCatalog struct {
	products map[string]domain.CatalogProduct
}

func (f fakeCatalog) GetProduct(_ context.Context, uuid string) (domain.CatalogProduct, error) {
	p, ok := f.products[uuid]
	if !ok {
		return domain.CatalogProduct{}, domain.ErrProductNotFound
	}
	return p, nil
}

func newCart(repo Repository) *Cart {
	cat := fakeCatalog{products: map[string]domain.CatalogProduct{
		"hat":  {UUID: "hat", Title: "Hat", Price: "9.99"},
		"shoe": {UUID: "shoe", Title: "Shoe", Price: "5.00"},
	}}
	return New(repo, cat)
}

func TestAddItemValidation(t *testing.T) {
	c := newCart(newMemRepo())

	if _, err := c.AddItem(context.Background(), "s1", "hat", 0); !errors.Is(err, domain.ErrInvalidQuantity) {
		t.Errorf("qty 0: err = %v, want ErrInvalidQuantity", err)
	}
	if _, err := c.AddItem(context.Background(), "s1", "ghost", 1); !errors.Is(err, domain.ErrProductNotFound) {
		t.Errorf("ghost product: err = %v, want ErrProductNotFound", err)
	}
}

func TestAddAndTotals(t *testing.T) {
	c := newCart(newMemRepo())
	ctx := context.Background()

	if _, err := c.AddItem(ctx, "s1", "hat", 2); err != nil {
		t.Fatal(err)
	}
	cart, err := c.AddItem(ctx, "s1", "shoe", 1)
	if err != nil {
		t.Fatal(err)
	}

	if len(cart.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(cart.Items))
	}
	// Сортировка по uuid: hat, shoe.
	if cart.Items[0].ProductUUID != "hat" || cart.Items[0].LineTotal != "19.98" {
		t.Errorf("hat line = %+v, want qty2 → 19.98", cart.Items[0])
	}
	if cart.Total != "24.98" {
		t.Errorf("total = %s, want 24.98", cart.Total)
	}
}

func TestGetCartSkipsVanishedProduct(t *testing.T) {
	repo := newMemRepo()
	repo.items["ghost"] = 3 // лежит в корзине, но в каталоге его уже нет
	repo.items["hat"] = 1
	c := newCart(repo)

	cart, err := c.GetCart(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if len(cart.Items) != 1 || cart.Items[0].ProductUUID != "hat" {
		t.Fatalf("want only hat, got %+v", cart.Items)
	}
	if cart.Total != "9.99" {
		t.Errorf("total = %s, want 9.99", cart.Total)
	}
}
