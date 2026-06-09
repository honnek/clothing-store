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

	catalogv1 "github.com/honnek/ranked-choice-shop/services/api/gen/catalog/v1"
	"github.com/honnek/ranked-choice-shop/services/internal/catalog/domain"
	"github.com/honnek/ranked-choice-shop/services/internal/catalog/usecase"
)

type fakeRepo struct{}

func (fakeRepo) ListProducts(context.Context, domain.ProductFilter, domain.Page) (domain.ProductList, error) {
	return domain.ProductList{
		Items: []domain.Product{{ID: 7, UUID: "u-1", Title: "Hat", Price: "9.99"}},
		Total: 1,
	}, nil
}
func (fakeRepo) GetProduct(_ context.Context, uuid string) (domain.Product, error) {
	if uuid == "good" {
		return domain.Product{ID: 7, UUID: "good", Title: "Hat", Price: "9.99"}, nil
	}
	return domain.Product{}, domain.ErrProductNotFound
}
func (fakeRepo) ListCategories(context.Context) ([]domain.Category, error) {
	return []domain.Category{{ID: 1, Title: "Hats", Slug: "hats"}}, nil
}

func newClient(t *testing.T) catalogv1.CatalogServiceClient {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	catalogv1.RegisterCatalogServiceServer(srv, NewServer(usecase.New(fakeRepo{})))
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
	return catalogv1.NewCatalogServiceClient(conn)
}

func TestListProducts(t *testing.T) {
	c := newClient(t)

	resp, err := c.ListProducts(context.Background(), &catalogv1.ListProductsRequest{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 || resp.Items[0].Uuid != "u-1" {
		t.Fatalf("got %+v", resp)
	}
}

func TestGetProduct(t *testing.T) {
	c := newClient(t)

	t.Run("found", func(t *testing.T) {
		resp, err := c.GetProduct(context.Background(), &catalogv1.GetProductRequest{Uuid: "good"})
		if err != nil {
			t.Fatal(err)
		}
		if resp.Product.Title != "Hat" {
			t.Fatalf("title = %q", resp.Product.Title)
		}
	})

	t.Run("missing -> NotFound", func(t *testing.T) {
		_, err := c.GetProduct(context.Background(), &catalogv1.GetProductRequest{Uuid: "nope"})
		if status.Code(err) != codes.NotFound {
			t.Fatalf("code = %s, want NotFound", status.Code(err))
		}
	})
}
