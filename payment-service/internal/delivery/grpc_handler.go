package delivery

import (
	"context"
	"database/sql"
	"ecommerce/pb"
	"ecommerce/payment-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentGrpcHandler struct {
	pb.UnimplementedPaymentServiceServer
	usecase domain.PaymentUseCase
}

func NewPaymentGrpcHandler(uc domain.PaymentUseCase) *PaymentGrpcHandler {
	return &PaymentGrpcHandler{usecase: uc}
}

func (h *PaymentGrpcHandler) GetPaymentStatus(ctx context.Context, req *pb.GetPaymentStatusRequest) (*pb.GetPaymentStatusResponse, error) {
	paymentStatus, err := h.usecase.GetStatus(ctx, req.GetOrderId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "payment record not found")
		}
		return nil, status.Error(codes.Internal, "database error")
	}
	return &pb.GetPaymentStatusResponse{Status: paymentStatus}, nil
}