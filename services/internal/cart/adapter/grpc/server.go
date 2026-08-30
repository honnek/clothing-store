package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cartv1 "github.com/honnek/lumewear-shop/services/api/gen/cart/v1"
	"github.com/honnek/lumewear-shop/services/internal/cart/domain"
	"github.com/honnek/lumewear-shop/services/internal/cart/usecase"
)

type Server struct {
	cartv1.UnimplementedCartServiceServer
	svc *usecase.Cart
}

func NewServer(svc *usecase.Cart) *Server {
	return &Server{svc: svc}
}

func (s *Server) GetCart(ctx context.Context, req *cartv1.GetCartRequest) (*cartv1.GetCartResponse, error) {
	cart, err := s.svc.GetCart(ctx, req.GetSessionId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &cartv1.GetCartResponse{Cart: toProto(cart)}, nil
}

func (s *Server) AddItem(ctx context.Context, req *cartv1.AddItemRequest) (*cartv1.AddItemResponse, error) {
	cart, err := s.svc.AddItem(ctx, req.GetSessionId(), req.GetProductUuid(), req.GetQuantity())
	if err != nil {
		return nil, toStatus(err)
	}
	return &cartv1.AddItemResponse{Cart: toProto(cart)}, nil
}

func (s *Server) UpdateItem(ctx context.Context, req *cartv1.UpdateItemRequest) (*cartv1.UpdateItemResponse, error) {
	cart, err := s.svc.UpdateItem(ctx, req.GetSessionId(), req.GetProductUuid(), req.GetQuantity())
	if err != nil {
		return nil, toStatus(err)
	}
	return &cartv1.UpdateItemResponse{Cart: toProto(cart)}, nil
}

func (s *Server) RemoveItem(ctx context.Context, req *cartv1.RemoveItemRequest) (*cartv1.RemoveItemResponse, error) {
	cart, err := s.svc.RemoveItem(ctx, req.GetSessionId(), req.GetProductUuid())
	if err != nil {
		return nil, toStatus(err)
	}
	return &cartv1.RemoveItemResponse{Cart: toProto(cart)}, nil
}

func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, "product not found")
	case errors.Is(err, domain.ErrInvalidQuantity):
		return status.Error(codes.InvalidArgument, "quantity must be positive")
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toProto(c domain.Cart) *cartv1.Cart {
	items := make([]*cartv1.CartItem, 0, len(c.Items))
	for _, it := range c.Items {
		items = append(items, &cartv1.CartItem{
			ProductUuid: it.ProductUUID,
			Title:       it.Title,
			UnitPrice:   it.UnitPrice,
			Quantity:    it.Quantity,
			LineTotal:   it.LineTotal,
		})
	}
	return &cartv1.Cart{SessionId: c.SessionID, Items: items, Total: c.Total}
}
