package transport

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc"

	orderv1 "github.com/honnek/lumewear-shop/services/api/gen/order/v1"
	grpcadapter "github.com/honnek/lumewear-shop/services/internal/order/adapter/grpc"
	"github.com/honnek/lumewear-shop/services/internal/order/domain"
	"github.com/honnek/lumewear-shop/services/internal/order/usecase"
)

type stubRepo struct{ gotKey string }

func (s *stubRepo) Checkout(_ context.Context, req domain.CheckoutRequest, _ []domain.CheckoutLine) (domain.Order, error) {
	s.gotKey = req.IdempotencyKey
	return domain.Order{ID: 42, Total: "19.98"}, nil
}
func (s *stubRepo) OrderByIdempotencyKey(context.Context, string) (domain.Order, error) {
	return domain.Order{}, domain.ErrOrderNotFound
}
func (s *stubRepo) Get(context.Context, int32) (domain.Order, error) { return domain.Order{}, nil }
func (s *stubRepo) List(context.Context, domain.OrderFilter, domain.Page) (domain.OrderList, error) {
	return domain.OrderList{}, nil
}
func (s *stubRepo) UpdateStatus(context.Context, int32, domain.Status) (domain.Order, error) {
	return domain.Order{}, nil
}

type stubCart struct{}

func (stubCart) Items(context.Context, string) (map[string]int32, error) {
	return map[string]int32{"hat": 2}, nil
}
func (stubCart) Clear(context.Context, string) error { return nil }

// Поднимает gRPC на свободном порту и gateway поверх него: заголовок Idempotency-Key
// должен дойти до хендлера, иначе REST-клиенту пришлось бы класть ключ в тело.
func newGateway(t *testing.T, repo *stubRepo) http.Handler {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	svc := usecase.New(repo, stubCart{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	gs := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(gs, grpcadapter.NewServer(svc))
	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(gs.Stop)

	gw, err := NewGateway(context.Background(), lis.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	return gw
}

func TestCheckoutOverREST(t *testing.T) {
	repo := &stubRepo{}
	gw := newGateway(t, repo)

	req := httptest.NewRequest(http.MethodPost, "/v1/checkout", strings.NewReader(`{"session_id":"s1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "k-42")
	rec := httptest.NewRecorder()

	gw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body)
	}
	if repo.gotKey != "k-42" {
		t.Errorf("idempotency key = %q, want k-42", repo.gotKey)
	}

	var body struct {
		Order struct {
			ID    int32  `json:"id"`
			Total string `json:"total"`
		} `json:"order"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Order.ID != 42 || body.Order.Total != "19.98" {
		t.Fatalf("got %+v", body.Order)
	}
}

func TestCheckoutWithoutKeyOverREST(t *testing.T) {
	gw := newGateway(t, &stubRepo{})

	req := httptest.NewRequest(http.MethodPost, "/v1/checkout", strings.NewReader(`{"session_id":"s1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	gw.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400; body = %s", rec.Code, rec.Body)
	}
}
