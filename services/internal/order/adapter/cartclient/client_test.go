package cartclient

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
)

type fakeCartServer struct {
	cartv1.UnimplementedCartServiceServer
	cart      *cartv1.Cart
	getErr    error
	clearedID string
}

func (f *fakeCartServer) GetCart(_ context.Context, _ *cartv1.GetCartRequest) (*cartv1.GetCartResponse, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &cartv1.GetCartResponse{Cart: f.cart}, nil
}

func (f *fakeCartServer) ClearCart(_ context.Context, req *cartv1.ClearCartRequest) (*cartv1.ClearCartResponse, error) {
	f.clearedID = req.GetSessionId()
	return &cartv1.ClearCartResponse{}, nil
}

func newClient(t *testing.T, srv *fakeCartServer) *Client {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)
	gs := grpc.NewServer()
	cartv1.RegisterCartServiceServer(gs, srv)
	go func() { _ = gs.Serve(lis) }()

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
		gs.Stop()
	})
	return New(conn)
}

func TestItems(t *testing.T) {
	c := newClient(t, &fakeCartServer{cart: &cartv1.Cart{
		SessionId: "s1",
		Items: []*cartv1.CartItem{
			{ProductUuid: "hat", Quantity: 2},
			{ProductUuid: "shoe", Quantity: 1},
		},
	}})

	items, err := c.Items(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items["hat"] != 2 || items["shoe"] != 1 {
		t.Fatalf("got %+v", items)
	}
}

func TestItemsPropagatesError(t *testing.T) {
	c := newClient(t, &fakeCartServer{getErr: status.Error(codes.Unavailable, "down")})

	if _, err := c.Items(context.Background(), "s1"); status.Code(err) != codes.Unavailable {
		t.Fatalf("err = %v, want Unavailable", err)
	}
}

func TestClear(t *testing.T) {
	srv := &fakeCartServer{cart: &cartv1.Cart{SessionId: "s1"}}
	c := newClient(t, srv)

	if err := c.Clear(context.Background(), "s1"); err != nil {
		t.Fatal(err)
	}
	if srv.clearedID != "s1" {
		t.Fatalf("cleared %q, want s1", srv.clearedID)
	}
}
