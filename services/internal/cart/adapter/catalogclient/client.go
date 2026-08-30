package catalogclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogv1 "github.com/honnek/lumewear-shop/services/api/gen/catalog/v1"
	"github.com/honnek/lumewear-shop/services/internal/cart/domain"
)

// Client — gRPC-клиент к catalog-сервису. Реализует usecase.Catalog: переводит
// gRPC-ответ в то, что нужно корзине, и NotFound — в доменную ошибку.
type Client struct {
	cli catalogv1.CatalogServiceClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{cli: catalogv1.NewCatalogServiceClient(conn)}
}

func (c *Client) GetProduct(ctx context.Context, uuid string) (domain.CatalogProduct, error) {
	resp, err := c.cli.GetProduct(ctx, &catalogv1.GetProductRequest{Uuid: uuid})
	if status.Code(err) == codes.NotFound {
		return domain.CatalogProduct{}, domain.ErrProductNotFound
	}
	if err != nil {
		return domain.CatalogProduct{}, fmt.Errorf("catalog get product: %w", err)
	}

	p := resp.GetProduct()
	return domain.CatalogProduct{UUID: p.GetUuid(), Title: p.GetTitle(), Price: p.GetPrice()}, nil
}
