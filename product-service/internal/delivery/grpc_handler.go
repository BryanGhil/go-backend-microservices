package delivery

import (
	"context"
	"database/sql"
	"ecommerce/pb"
	"ecommerce/product-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductGrpcHandler struct {
	pb.UnimplementedProductServiceServer
	usecase domain.ProductUseCase
}

func NewProductGrpcHandler(uc domain.ProductUseCase) *ProductGrpcHandler {
	return &ProductGrpcHandler{usecase: uc}
}

func (h *ProductGrpcHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, err := h.usecase.GetProduct(ctx, req.GetId())
	if err != nil {
		// Translation: If DB says no rows, return gRPC NotFound
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.GetProductResponse{Id: product.ID, Name: product.Name, Price: float32(product.Price)}, nil
}

func (h *ProductGrpcHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	id, err := h.usecase.CreateProduct(ctx, req.GetName(), float64(req.GetPrice()))
	if err != nil {
		return nil, err
	}

	return &pb.CreateProductResponse{Id: id}, nil
}

func (h *ProductGrpcHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	err := h.usecase.UpdateProduct(ctx, req.GetId(), req.GetName(), float64(req.GetPrice()))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "cannot update, product not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateProductResponse{Success: true}, nil
}

func (h *ProductGrpcHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	err := h.usecase.DeleteProduct(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteProductResponse{Success: true}, nil
}

func (h *ProductGrpcHandler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, err := h.usecase.GetAllProducts(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert Go slice to Protobuf repeated field
	var pbProducts []*pb.GetProductResponse
	for _, p := range products {
		pbProducts = append(pbProducts, &pb.GetProductResponse{
			Id:    p.ID,
			Name:  p.Name,
			Price: float32(p.Price),
		})
	}

	return &pb.ListProductsResponse{Products: pbProducts}, nil
}