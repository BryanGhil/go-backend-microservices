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
	id, err := h.usecase.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to register user")
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
	id, err := h.usecase.VerifySession(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired session")
	}
	return &pb.VerifySessionResponse{UserId: id}, nil
}

func (h *UserGrpcHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	err := h.usecase.UpdateProfile(ctx, req.GetUserId(), req.GetFullName(), req.GetPhone(), req.GetAddress())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update profile")
	}
	return &pb.UpdateProfileResponse{Success: true}, nil
}