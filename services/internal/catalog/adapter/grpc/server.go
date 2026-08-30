package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogv1 "github.com/honnek/lumewear-shop/services/api/gen/catalog/v1"
	"github.com/honnek/lumewear-shop/services/internal/catalog/domain"
	"github.com/honnek/lumewear-shop/services/internal/catalog/usecase"
)

type Server struct {
	catalogv1.UnimplementedCatalogServiceServer
	svc *usecase.Catalog
}

func NewServer(svc *usecase.Catalog) *Server {
	return &Server{svc: svc}
}

func (s *Server) ListProducts(ctx context.Context, req *catalogv1.ListProductsRequest) (*catalogv1.ListProductsResponse, error) {
	list, err := s.svc.ListProducts(ctx,
		domain.ProductFilter{
			CategoryID: req.CategoryId,
			Published:  req.Published,
			Search:     req.Search,
		},
		domain.Page{Limit: req.Limit, Offset: req.Offset},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	items := make([]*catalogv1.Product, 0, len(list.Items))
	for i := range list.Items {
		items = append(items, toProto(list.Items[i]))
	}
	return &catalogv1.ListProductsResponse{Items: items, Total: list.Total}, nil
}

func (s *Server) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	p, err := s.svc.GetProduct(ctx, req.GetUuid())
	if errors.Is(err, domain.ErrProductNotFound) {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &catalogv1.GetProductResponse{Product: toProto(p)}, nil
}

func (s *Server) ListCategories(ctx context.Context, _ *catalogv1.ListCategoriesRequest) (*catalogv1.ListCategoriesResponse, error) {
	cats, err := s.svc.ListCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	out := make([]*catalogv1.Category, 0, len(cats))
	for _, c := range cats {
		out = append(out, &catalogv1.Category{Id: c.ID, Title: c.Title, Slug: c.Slug})
	}
	return &catalogv1.ListCategoriesResponse{Items: out}, nil
}

func toProto(p domain.Product) *catalogv1.Product {
	return &catalogv1.Product{
		Uuid:        p.UUID,
		Title:       p.Title,
		Price:       p.Price,
		Quality:     p.Quality,
		Description: p.Description,
		Slug:        p.Slug,
		CategoryId:  p.CategoryID,
		IsPublished: p.IsPublished,
	}
}
