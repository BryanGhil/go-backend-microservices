package delivery

import (
	"context"
	"database/sql"
	"ecommerce/order-service/internal/domain"
	"ecommerce/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGrpcHandler struct {
	pb.UnimplementedOrderServiceServer
	usecase domain.OrderUseCase
}

// Constructor for the main.go wiring
func NewOrderGrpcHandler(uc domain.OrderUseCase) *OrderGrpcHandler {
	return &OrderGrpcHandler{usecase: uc}
}

// Handles the Checkout process
func (h *OrderGrpcHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	// Call the UseCase
	orderID, err := h.usecase.Checkout(ctx, req.GetUserId(), req.GetProductId(), float64(req.GetAmount()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to process checkout")
	}

	return &pb.CreateOrderResponse{OrderId: orderID}, nil
}

// Handles checking the status of the order (Pending, Completed, Cancelled)
func (h *OrderGrpcHandler) GetOrderStatus(ctx context.Context, req *pb.GetOrderStatusRequest) (*pb.GetOrderStatusResponse, error) {
	orderStatus, err := h.usecase.GetStatus(ctx, req.GetOrderId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "failed to retrieve order status")
	}

	return &pb.GetOrderStatusResponse{Status: orderStatus}, nil
}