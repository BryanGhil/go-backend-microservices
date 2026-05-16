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
	// Map Protobuf items to Domain items
	var items []domain.CheckoutItem
	for _, reqItem := range req.GetItems() {
		items = append(items, domain.CheckoutItem{
			ProductID: reqItem.GetProductId(),
			Quantity:  int(reqItem.GetQuantity()),
		})
	}

	// Ensure there is at least one item to checkout
	if len(items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "cannot checkout empty cart")
	}

	// Call UseCase
	correlationID, err := h.usecase.Checkout(ctx, req.GetUserId(), items)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to process checkout")
	}

	// Return the CorrelationID to the API Gateway so it can pass it to the Payment Service!
	return &pb.CreateOrderResponse{CorrelationId: correlationID}, nil
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

func (h *OrderGrpcHandler) GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.GetUserOrdersResponse, error) {
	orders, err := h.usecase.GetUserOrders(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch user orders")
	}

	var pbOrders []*pb.OrderItemResponse
	for _, o := range orders {
		pbOrders = append(pbOrders, &pb.OrderItemResponse{
			Id:            o.ID,
			Amount:        o.Amount,
			Status:        o.Status,
			CorrelationId: o.CorrelationID,
		})
	}

	return &pb.GetUserOrdersResponse{Orders: pbOrders}, nil
}