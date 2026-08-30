package grpc

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	cartv1 "github.com/honnek/lumewear-shop/services/api/gen/cart/v1"
	"github.com/honnek/lumewear-shop/services/internal/cart/domain"
	"github.com/honnek/lumewear-shop/services/internal/cart/usecase"
)

type memRepo struct{ items map[string]int32 }

func (m *memRepo) Add(_ context.Context, _, uuid string, delta int32) error {
	m.items[uuid] += delta
	return nil
}
func (m *memRepo) SetQty(_ context.Context, _, uuid string, qty int32) error {
	m.items[uuid] = qty
	return nil
}
func (m *memRepo) Remove(_ context.Context, _, uuid string) error { delete(m.items, uuid); return nil }
func (m *memRepo) Clear(_ context.Context, _ string) error {
	m.items = map[string]int32{}
	return nil
}
func (m *memRepo) Items(_ context.Context, _ string) (map[string]int32, error) {
	return m.items, nil
}

type fakeCatalog struct{}

func (fakeCatalog) GetProduct(_ context.Context, uuid string) (domain.CatalogProduct, error) {
	if uuid == "hat" {
		return domain.CatalogProduct{UUID: "hat", Title: "Hat", Price: "9.99"}, nil
	}
	return domain.CatalogProduct{}, domain.ErrProductNotFound
}

func newClient(t *testing.T) cartv1.CartServiceClient {
	t.Helper()

	svc := usecase.New(&memRepo{items: map[string]int32{}}, fakeCatalog{})
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	cartv1.RegisterCartServiceServer(srv, NewServer(svc))
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
		srv.Stop()
	})
	return cartv1.NewCartServiceClient(conn)
}

func TestAddItem(t *testing.T) {
	c := newClient(t)

	resp, err := c.AddItem(context.Background(), &cartv1.AddItemRequest{SessionId: "s1", ProductUuid: "hat", Quantity: 2})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Cart.Total != "19.98" || len(resp.Cart.Items) != 1 {
		t.Fatalf("got %+v", resp.Cart)
	}
}

func TestAddItemErrors(t *testing.T) {
	c := newClient(t)

	_, err := c.AddItem(context.Background(), &cartv1.AddItemRequest{SessionId: "s1", ProductUuid: "ghost", Quantity: 1})
	if status.Code(err) != codes.NotFound {
		t.Errorf("ghost: code = %s, want NotFound", status.Code(err))
	}

	_, err = c.AddItem(context.Background(), &cartv1.AddItemRequest{SessionId: "s1", ProductUuid: "hat", Quantity: 0})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("qty 0: code = %s, want InvalidArgument", status.Code(err))
	}
}

func TestClearCart(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	if _, err := c.AddItem(ctx, &cartv1.AddItemRequest{SessionId: "s1", ProductUuid: "hat", Quantity: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ClearCart(ctx, &cartv1.ClearCartRequest{SessionId: "s1"}); err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetCart(ctx, &cartv1.GetCartRequest{SessionId: "s1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Cart.Items) != 0 || resp.Cart.Total != "0.00" {
		t.Fatalf("got %+v", resp.Cart)
	}
}
