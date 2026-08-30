package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/honnek/lumewear-shop/services/internal/order/domain"
)

type fakeRepo struct {
	gotLines []domain.CheckoutLine
	gotPage  domain.Page
	err      error
	byKey    map[string]domain.Order
}

func (f *fakeRepo) Checkout(_ context.Context, _ domain.CheckoutRequest, lines []domain.CheckoutLine) (domain.Order, error) {
	f.gotLines = lines
	return domain.Order{ID: 1}, f.err
}
func (f *fakeRepo) OrderByIdempotencyKey(_ context.Context, key string) (domain.Order, error) {
	o, ok := f.byKey[key]
	if !ok {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return o, nil
}
func (f *fakeRepo) Get(context.Context, int32) (domain.Order, error) { return domain.Order{}, nil }
func (f *fakeRepo) List(_ context.Context, _ domain.OrderFilter, p domain.Page) (domain.OrderList, error) {
	f.gotPage = p
	return domain.OrderList{}, nil
}
func (f *fakeRepo) UpdateStatus(context.Context, int32, domain.Status) (domain.Order, error) {
	return domain.Order{}, nil
}

type fakeCart struct {
	items    map[string]int32
	cleared  bool
	clearErr error
}

func (f *fakeCart) Items(context.Context, string) (map[string]int32, error) { return f.items, nil }
func (f *fakeCart) Clear(context.Context, string) error {
	f.cleared = true
	return f.clearErr
}

func newOrder(repo Repository, cart Cart) *Order {
	return New(repo, cart, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func req() domain.CheckoutRequest {
	return domain.CheckoutRequest{SessionID: "s1", IdempotencyKey: "k1"}
}

func TestCheckoutRejectsBadInput(t *testing.T) {
	t.Run("no idempotency key", func(t *testing.T) {
		o := newOrder(&fakeRepo{}, &fakeCart{items: map[string]int32{"u": 1}})

		_, err := o.Checkout(context.Background(), domain.CheckoutRequest{SessionID: "s1"})
		if !errors.Is(err, domain.ErrIdempotencyKeyRequired) {
			t.Fatalf("err = %v, want ErrIdempotencyKeyRequired", err)
		}
	})

	t.Run("empty cart", func(t *testing.T) {
		o := newOrder(&fakeRepo{}, &fakeCart{items: map[string]int32{}})

		if _, err := o.Checkout(context.Background(), req()); !errors.Is(err, domain.ErrEmptyCart) {
			t.Fatalf("err = %v, want ErrEmptyCart", err)
		}
	})

	t.Run("cart with only junk quantities", func(t *testing.T) {
		o := newOrder(&fakeRepo{}, &fakeCart{items: map[string]int32{"u": 0, "v": -2}})

		if _, err := o.Checkout(context.Background(), req()); !errors.Is(err, domain.ErrEmptyCart) {
			t.Fatalf("err = %v, want ErrEmptyCart", err)
		}
	})
}

// Ретрай checkout-а приходит уже с пустой корзиной: заказ по ключу должен вернуться
// прежний, иначе клиент, не дождавшийся первого ответа, увидит «корзина пуста».
func TestCheckoutRetryAfterCartCleared(t *testing.T) {
	repo := &fakeRepo{byKey: map[string]domain.Order{"k1": {ID: 42}}}
	o := newOrder(repo, &fakeCart{items: map[string]int32{}})

	order, err := o.Checkout(context.Background(), req())
	if err != nil {
		t.Fatal(err)
	}
	if order.ID != 42 {
		t.Fatalf("order id = %d, want 42", order.ID)
	}
}

func TestCheckoutBuildsStableLines(t *testing.T) {
	repo := &fakeRepo{}
	cart := &fakeCart{items: map[string]int32{"c": 1, "a": 2, "b": 3}}

	if _, err := newOrder(repo, cart).Checkout(context.Background(), req()); err != nil {
		t.Fatal(err)
	}

	want := []string{"a", "b", "c"}
	if len(repo.gotLines) != len(want) {
		t.Fatalf("lines = %+v", repo.gotLines)
	}
	for i, uuid := range want {
		if repo.gotLines[i].ProductUUID != uuid {
			t.Fatalf("lines = %+v, want sorted by uuid", repo.gotLines)
		}
	}
	if !cart.cleared {
		t.Error("cart must be cleared after checkout")
	}
}

// Корзина уже не нужна: заказ создан, и провал её очистки не должен превращаться
// в ошибку для клиента.
func TestCheckoutSurvivesCartClearFailure(t *testing.T) {
	cart := &fakeCart{items: map[string]int32{"u": 1}, clearErr: errors.New("redis down")}

	order, err := newOrder(&fakeRepo{}, cart).Checkout(context.Background(), req())
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if order.ID != 1 {
		t.Errorf("order = %+v", order)
	}
}

func TestCheckoutKeepsCartOnFailure(t *testing.T) {
	cart := &fakeCart{items: map[string]int32{"u": 1}}
	repo := &fakeRepo{err: domain.ErrInsufficientStock}

	if _, err := newOrder(repo, cart).Checkout(context.Background(), req()); !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("err = %v", err)
	}
	if cart.cleared {
		t.Error("cart must survive a failed checkout")
	}
}

func TestListClampsPagination(t *testing.T) {
	tests := []struct {
		name       string
		in         domain.Page
		wantLimit  int32
		wantOffset int32
	}{
		{"zero limit -> default", domain.Page{}, defaultLimit, 0},
		{"over max -> default", domain.Page{Limit: 500, Offset: 5}, defaultLimit, 5},
		{"valid stays", domain.Page{Limit: 30, Offset: 10}, 30, 10},
		{"negative offset -> 0", domain.Page{Limit: 10, Offset: -1}, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			if _, err := newOrder(repo, &fakeCart{}).List(context.Background(), domain.OrderFilter{}, tt.in); err != nil {
				t.Fatal(err)
			}
			if repo.gotPage.Limit != tt.wantLimit || repo.gotPage.Offset != tt.wantOffset {
				t.Errorf("page = %+v, want {%d %d}", repo.gotPage, tt.wantLimit, tt.wantOffset)
			}
		})
	}
}

func TestUpdateStatusRejectsUnknownStatus(t *testing.T) {
	o := newOrder(&fakeRepo{}, &fakeCart{})

	if _, err := o.UpdateStatus(context.Background(), 1, domain.Status(9)); !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Fatalf("err = %v, want ErrInvalidStatusTransition", err)
	}
}
