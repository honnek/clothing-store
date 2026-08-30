package grpc

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	orderv1 "github.com/honnek/lumewear-shop/services/api/gen/order/v1"
	"github.com/honnek/lumewear-shop/services/internal/order/domain"
	"github.com/honnek/lumewear-shop/services/internal/order/usecase"
)

type stubRepo struct {
	gotKey  string
	order   domain.Order
	err     error
	listErr error
}

func (s *stubRepo) Checkout(_ context.Context, req domain.CheckoutRequest, _ []domain.CheckoutLine) (domain.Order, error) {
	s.gotKey = req.IdempotencyKey
	return s.order, s.err
}

func (s *stubRepo) OrderByIdempotencyKey(context.Context, string) (domain.Order, error) {
	return domain.Order{}, domain.ErrOrderNotFound
}

func (s *stubRepo) Get(context.Context, int32) (domain.Order, error) {
	return s.order, s.err
}

func (s *stubRepo) List(context.Context, domain.OrderFilter, domain.Page) (domain.OrderList, error) {
	if s.listErr != nil {
		return domain.OrderList{}, s.listErr
	}
	return domain.OrderList{Items: []domain.Order{s.order}, Total: 1}, nil
}

func (s *stubRepo) UpdateStatus(context.Context, int32, domain.Status) (domain.Order, error) {
	return s.order, s.err
}

type stubCart struct{ items map[string]int32 }

func (c stubCart) Items(context.Context, string) (map[string]int32, error) { return c.items, nil }
func (c stubCart) Clear(context.Context, string) error                     { return nil }

func sampleOrder() domain.Order {
	return domain.Order{
		ID:        7,
		Status:    domain.StatusCreated,
		Total:     "19.98",
		CreatedAt: time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC),
		Items: []domain.Item{
			{ProductID: 1, ProductUUID: "hat", Title: "Hat", UnitPrice: "9.99", Quantity: 2, LineTotal: "19.98"},
		},
	}
}

func newClient(t *testing.T, repo *stubRepo) orderv1.OrderServiceClient {
	t.Helper()

	svc := usecase.New(repo, stubCart{items: map[string]int32{"hat": 2}},
		slog.New(slog.NewTextHandler(io.Discard, nil)))

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(srv, NewServer(svc))
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
	return orderv1.NewOrderServiceClient(conn)
}

func TestCheckout(t *testing.T) {
	repo := &stubRepo{order: sampleOrder()}
	c := newClient(t, repo)

	resp, err := c.Checkout(context.Background(), &orderv1.CheckoutRequest{SessionId: "s1", IdempotencyKey: "k1"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.gotKey != "k1" {
		t.Errorf("idempotency key = %q, want k1", repo.gotKey)
	}

	got := resp.GetOrder()
	if got.GetId() != 7 || got.GetTotal() != "19.98" || len(got.GetItems()) != 1 {
		t.Fatalf("got %+v", got)
	}
	if got.GetStatus() != orderv1.OrderStatus_ORDER_STATUS_CREATED {
		t.Errorf("status = %s", got.GetStatus())
	}
	if got.GetCreatedAt() != "2026-08-30T12:00:00Z" {
		t.Errorf("created_at = %q", got.GetCreatedAt())
	}
}

// Ключ идемпотентности REST-клиент шлёт заголовком; gateway кладёт его в метаданные.
func TestCheckoutTakesKeyFromMetadata(t *testing.T) {
	repo := &stubRepo{order: sampleOrder()}
	c := newClient(t, repo)

	ctx := metadata.AppendToOutgoingContext(context.Background(), IdempotencyKeyHeader, "from-header")
	if _, err := c.Checkout(ctx, &orderv1.CheckoutRequest{SessionId: "s1"}); err != nil {
		t.Fatal(err)
	}
	if repo.gotKey != "from-header" {
		t.Fatalf("idempotency key = %q, want from-header", repo.gotKey)
	}
}

func TestCheckoutErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want codes.Code
	}{
		{"insufficient stock", &domain.StockError{ProductUUID: "hat", Requested: 3, Available: 1}, codes.FailedPrecondition},
		{"product gone", domain.ErrProductNotFound, codes.FailedPrecondition},
		{"unknown", context.DeadlineExceeded, codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, &stubRepo{err: tt.err})

			_, err := c.Checkout(context.Background(), &orderv1.CheckoutRequest{SessionId: "s1", IdempotencyKey: "k1"})
			if status.Code(err) != tt.want {
				t.Fatalf("code = %s, want %s", status.Code(err), tt.want)
			}
		})
	}
}

func TestCheckoutWithoutKey(t *testing.T) {
	c := newClient(t, &stubRepo{order: sampleOrder()})

	_, err := c.Checkout(context.Background(), &orderv1.CheckoutRequest{SessionId: "s1"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %s, want InvalidArgument", status.Code(err))
	}
}

func TestGetOrderNotFound(t *testing.T) {
	c := newClient(t, &stubRepo{err: domain.ErrOrderNotFound})

	_, err := c.GetOrder(context.Background(), &orderv1.GetOrderRequest{Id: 404})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("code = %s, want NotFound", status.Code(err))
	}
}

func TestListOrders(t *testing.T) {
	c := newClient(t, &stubRepo{order: sampleOrder()})

	denied := orderv1.OrderStatus_ORDER_STATUS_DENIED
	resp, err := c.ListOrders(context.Background(), &orderv1.ListOrdersRequest{Status: &denied, Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetTotal() != 1 || len(resp.GetItems()) != 1 {
		t.Fatalf("got %+v", resp)
	}
}

func TestUpdateOrderStatus(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		order := sampleOrder()
		order.Status = domain.StatusProcessed
		c := newClient(t, &stubRepo{order: order})

		resp, err := c.UpdateOrderStatus(context.Background(), &orderv1.UpdateOrderStatusRequest{
			Id:     7,
			Status: orderv1.OrderStatus_ORDER_STATUS_PROCESSED,
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp.GetOrder().GetStatus() != orderv1.OrderStatus_ORDER_STATUS_PROCESSED {
			t.Fatalf("status = %s", resp.GetOrder().GetStatus())
		}
	})

	t.Run("forbidden transition", func(t *testing.T) {
		c := newClient(t, &stubRepo{err: &domain.TransitionError{From: domain.StatusDelivered, To: domain.StatusCreated}})

		_, err := c.UpdateOrderStatus(context.Background(), &orderv1.UpdateOrderStatusRequest{Id: 7})
		if status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("code = %s, want FailedPrecondition", status.Code(err))
		}
	})

	t.Run("unknown status value", func(t *testing.T) {
		c := newClient(t, &stubRepo{order: sampleOrder()})

		_, err := c.UpdateOrderStatus(context.Background(), &orderv1.UpdateOrderStatusRequest{Id: 7, Status: 42})
		if status.Code(err) != codes.FailedPrecondition {
			t.Fatalf("code = %s, want FailedPrecondition", status.Code(err))
		}
	})
}
