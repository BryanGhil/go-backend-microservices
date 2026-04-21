package delivery

import (
	"context"
	"database/sql"
	"ecommerce/inventory-service/internal/domain"
	"ecommerce/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InventoryGrpcHandler struct {
	pb.UnimplementedInventoryServiceServer
	usecase domain.InventoryUseCase
}

// Constructor for the main.go wiring
func NewInventoryGrpcHandler(uc domain.InventoryUseCase) *InventoryGrpcHandler {
	return &InventoryGrpcHandler{usecase: uc}
}

// Handles fetching current stock levels
func (h *InventoryGrpcHandler) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	stock, err := h.usecase.GetStock(ctx, req.GetProductId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "inventory record not found for this product")
		}
		return nil, status.Error(codes.Internal, "failed to retrieve stock")
	}

	return &pb.GetStockResponse{Stock: stock}, nil
}

// Handles manual stock additions by an admin/seller
func (h *InventoryGrpcHandler) AddStock(ctx context.Context, req *pb.AddStockRequest) (*pb.AddStockResponse, error) {
	err := h.usecase.AddStock(ctx, req.GetProductId(), req.GetQuantity())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add stock")
	}

	return &pb.AddStockResponse{Success: true}, nil
}