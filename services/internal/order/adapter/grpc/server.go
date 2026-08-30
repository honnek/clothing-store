package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	orderv1 "github.com/honnek/lumewear-shop/services/api/gen/order/v1"
	"github.com/honnek/lumewear-shop/services/internal/order/domain"
	"github.com/honnek/lumewear-shop/services/internal/order/usecase"
)

// IdempotencyKeyHeader — HTTP-заголовок, из которого gateway кладёт ключ в метаданные.
const IdempotencyKeyHeader = "idempotency-key"

type Server struct {
	orderv1.UnimplementedOrderServiceServer
	svc *usecase.Order
}

func NewServer(svc *usecase.Order) *Server {
	return &Server{svc: svc}
}

func (s *Server) Checkout(ctx context.Context, req *orderv1.CheckoutRequest) (*orderv1.CheckoutResponse, error) {
	key := req.GetIdempotencyKey()
	if key == "" {
		key = keyFromMetadata(ctx)
	}

	order, err := s.svc.Checkout(ctx, domain.CheckoutRequest{
		SessionID:      req.GetSessionId(),
		IdempotencyKey: key,
		OwnerID:        req.OwnerId,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &orderv1.CheckoutResponse{Order: toProto(order)}, nil
}

func (s *Server) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	order, err := s.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &orderv1.GetOrderResponse{Order: toProto(order)}, nil
}

func (s *Server) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	var filter domain.OrderFilter
	filter.OwnerID = req.OwnerId
	if req.Status != nil {
		st := domain.Status(req.GetStatus())
		filter.Status = &st
	}

	list, err := s.svc.List(ctx, filter, domain.Page{Limit: req.GetLimit(), Offset: req.GetOffset()})
	if err != nil {
		return nil, toStatus(err)
	}

	items := make([]*orderv1.Order, 0, len(list.Items))
	for _, o := range list.Items {
		items = append(items, toProto(o))
	}
	return &orderv1.ListOrdersResponse{Items: items, Total: list.Total}, nil
}

func (s *Server) UpdateOrderStatus(ctx context.Context, req *orderv1.UpdateOrderStatusRequest) (*orderv1.UpdateOrderStatusResponse, error) {
	order, err := s.svc.UpdateStatus(ctx, req.GetId(), domain.Status(req.GetStatus()))
	if err != nil {
		return nil, toStatus(err)
	}
	return &orderv1.UpdateOrderStatusResponse{Order: toProto(order)}, nil
}

func keyFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(IdempotencyKeyHeader)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Нехватка остатка и запрещённый переход — не ошибка запроса, а состояние системы,
// поэтому FailedPrecondition: клиенту нет смысла ретраить с тем же телом.
func toStatus(err error) error {
	var stock *domain.StockError
	if errors.As(err, &stock) {
		return status.Error(codes.FailedPrecondition, stock.Error())
	}

	var transition *domain.TransitionError
	if errors.As(err, &transition) {
		return status.Error(codes.FailedPrecondition, transition.Error())
	}

	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		return status.Error(codes.NotFound, "order not found")
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.FailedPrecondition, "product from cart is no longer available")
	case errors.Is(err, domain.ErrEmptyCart):
		return status.Error(codes.FailedPrecondition, "cart is empty")
	case errors.Is(err, domain.ErrIdempotencyKeyRequired):
		return status.Error(codes.InvalidArgument, "idempotency key is required")
	case errors.Is(err, domain.ErrInvalidStatusTransition):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toProto(o domain.Order) *orderv1.Order {
	items := make([]*orderv1.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, &orderv1.OrderItem{
			ProductUuid: it.ProductUUID,
			Title:       it.Title,
			UnitPrice:   it.UnitPrice,
			Quantity:    it.Quantity,
			LineTotal:   it.LineTotal,
		})
	}

	return &orderv1.Order{
		Id:        o.ID,
		OwnerId:   o.OwnerID,
		Status:    orderv1.OrderStatus(o.Status),
		Total:     o.Total,
		CreatedAt: o.CreatedAt.Format(time.RFC3339),
		UpdatedAt: o.UpdatedAt.Format(time.RFC3339),
		Items:     items,
	}
}
