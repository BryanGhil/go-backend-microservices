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
        if err == sql.ErrNoRows {
            return nil, status.Error(codes.NotFound, "product not found")
        }
        return nil, status.Error(codes.Internal, "internal server error")
    }

    // Use the new nested Product message
    return &pb.GetProductResponse{
        Product: toPbProduct(product),
    }, nil
}

func (h *ProductGrpcHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
    // Map the incoming gRPC request to your Domain entity
    domainProduct := &domain.Product{
        SellerID:    req.GetSellerId(),
        Name:        req.GetName(),
        Description: req.GetDescription(),
        Category:    req.GetCategory(),
        Price:       req.GetPrice(),
        ImageURL:    req.GetImageUrl(),
    }

    id, err := h.usecase.CreateProduct(ctx, domainProduct)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.CreateProductResponse{Id: id}, nil
}

func (h *ProductGrpcHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
    domainProduct := &domain.Product{
        ID:          req.GetId(),
        Name:        req.GetName(),
        Description: req.GetDescription(),
        Category:    req.GetCategory(),
        Price:       req.GetPrice(),
        ImageURL:    req.GetImageUrl(),
        IsActive:    req.GetIsActive(),
    }

    err := h.usecase.UpdateProduct(ctx, domainProduct)
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
    // Pass the new pagination parameters
    products, totalCount, err := h.usecase.ListProducts(
        ctx, 
        req.GetLimit(), 
        req.GetOffset(), 
        req.GetSellerId(), 
        req.GetCategory(),
    )
    
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    var pbProducts []*pb.Product
    for _, p := range products {
        pbProducts = append(pbProducts, toPbProduct(p))
    }

    return &pb.ListProductsResponse{
        Products:   pbProducts,
        TotalCount: totalCount,
    }, nil
}

// --- Helper Function to map Domain -> Protobuf ---
func toPbProduct(p *domain.Product) *pb.Product {
    if p == nil {
        return nil
    }
    return &pb.Product{
        Id:          p.ID,
        SellerId:    p.SellerID,
        Name:        p.Name,
        Description: p.Description,
        Category:    p.Category,
        Price:       p.Price, // Now a float64 (double in proto)
        ImageUrl:    p.ImageURL,
        IsActive:    p.IsActive,
    }
}