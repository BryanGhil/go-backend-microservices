package delivery

import (
	"context"
	"ecommerce/pb"
	"ecommerce/user-service/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserGrpcHandler struct {
	pb.UnimplementedUserServiceServer
	usecase domain.UserUseCase
}

func NewUserGrpcHandler(uc domain.UserUseCase) *UserGrpcHandler {
	return &UserGrpcHandler{usecase: uc}
}

func (h *UserGrpcHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Map proto request to domain struct
	domainUser := &domain.User{
		Email:    req.GetEmail(),
		FullName: req.GetFullName(),
		Role:     req.GetRole(),
	}

	// Add seller profile if applicable
	if req.GetRole() == "seller" {
		domainUser.SellerProfile = &domain.SellerProfile{
			ShopName: req.GetShopName(),
		}
	}

	id, err := h.usecase.Register(ctx, domainUser, req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RegisterResponse{UserId: id}, nil
}

func (h *UserGrpcHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := h.usecase.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return &pb.LoginResponse{Token: token}, nil
}

func (h *UserGrpcHandler) VerifySession(ctx context.Context, req *pb.VerifySessionRequest) (*pb.VerifySessionResponse, error) {
	id, role, err := h.usecase.VerifySession(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired session")
	}
	
	// Now returning both ID and Role to the API Gateway!
	return &pb.VerifySessionResponse{
		UserId:  id,
		Role:    role,
		IsValid: true,
	}, nil
}

func (h *UserGrpcHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	domainUser := &domain.User{
		ID:       req.GetUserId(),
		FullName: req.GetFullName(),
		Phone:    req.GetPhone(),
		Address:  req.GetAddress(),
	}

	// Pass seller updates if they exist
	if req.GetShopName() != "" || req.GetShopDescription() != "" {
		domainUser.SellerProfile = &domain.SellerProfile{
			UserID:          req.GetUserId(),
			ShopName:        req.GetShopName(),
			ShopDescription: req.GetShopDescription(),
		}
	}

	err := h.usecase.UpdateProfile(ctx, domainUser)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update profile")
	}
	return &pb.UpdateProfileResponse{Success: true}, nil
}