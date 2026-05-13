package delivery

import (
	"context"

	"ecommerce/cart-service/internal/domain"
	"ecommerce/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartGrpcHandler struct {
	pb.UnimplementedCartServiceServer
	uc domain.CartUseCase
}

func NewCartGrpcHandler(uc domain.CartUseCase) *CartGrpcHandler {
	return &CartGrpcHandler{
		uc: uc,
	}
}

func (h *CartGrpcHandler) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.GetCartResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	cartItems, err := h.uc.GetCart(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get cart")
	}

	var pbItems []*pb.CartItem
	for _, item := range cartItems {
		pbItems = append(pbItems, &pb.CartItem{
			ProductId:      item.ProductID,
			Quantity:       item.Quantity,
			Name:           item.Name,
			Price:          item.Price,
			Subtotal:       item.Subtotal,
			ImageUrl:       item.ImageURL,
			SellerShopName: item.SellerShopName,
			SellerId:       item.SellerID,
		})
	}

	return &pb.GetCartResponse{Items: pbItems}, nil
}

func (h *CartGrpcHandler) AddToCart(ctx context.Context, req *pb.AddToCartRequest) (*pb.AddToCartResponse, error) {
	if req.GetUserId() == 0 || req.GetProductId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID and product ID are required")
	}

	err := h.uc.AddToCart(ctx, req.GetUserId(), req.GetProductId(), req.GetQuantity())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add item to cart")
	}

	return &pb.AddToCartResponse{Success: true}, nil
}

func (h *CartGrpcHandler) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.RemoveItemResponse, error) {
	if req.GetUserId() == 0 || req.GetProductId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID and product ID are required")
	}

	err := h.uc.RemoveItem(ctx, req.GetUserId(), req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to remove item from cart")
	}

	return &pb.RemoveItemResponse{Success: true}, nil
}

func (h *CartGrpcHandler) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*pb.ClearCartResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	err := h.uc.ClearCart(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to clear cart")
	}

	return &pb.ClearCartResponse{Success: true}, nil
}

func (h *CartGrpcHandler) GetCartCount(ctx context.Context, req *pb.GetCartCountRequest) (*pb.GetCartCountResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	count, err := h.uc.GetCartCount(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get cart count")
	}

	return &pb.GetCartCountResponse{Count: count}, nil
}
