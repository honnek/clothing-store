package cartclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	cartv1 "github.com/honnek/lumewear-shop/services/api/gen/cart/v1"
)

// Client — gRPC-клиент к cart-service, реализующий usecase.Cart. Из корзины заказу
// нужны только количества: цену он всё равно берёт из товара под блокировкой.
type Client struct {
	cli cartv1.CartServiceClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{cli: cartv1.NewCartServiceClient(conn)}
}

func (c *Client) Items(ctx context.Context, session string) (map[string]int32, error) {
	resp, err := c.cli.GetCart(ctx, &cartv1.GetCartRequest{SessionId: session})
	if err != nil {
		return nil, fmt.Errorf("cart get: %w", err)
	}

	items := resp.GetCart().GetItems()
	out := make(map[string]int32, len(items))
	for _, it := range items {
		out[it.GetProductUuid()] = it.GetQuantity()
	}
	return out, nil
}

func (c *Client) Clear(ctx context.Context, session string) error {
	if _, err := c.cli.ClearCart(ctx, &cartv1.ClearCartRequest{SessionId: session}); err != nil {
		return fmt.Errorf("cart clear: %w", err)
	}
	return nil
}
