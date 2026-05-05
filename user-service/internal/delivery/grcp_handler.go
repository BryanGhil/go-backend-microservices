package delivery

import (
	"context"
	"ecommerce/pb"
	"ecommerce/user-service/internal/domain"
	"time"

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
	domainUser := &domain.User{
		Email:    req.GetEmail(),
		FullName: req.GetFullName(),
		Role:     req.GetRole(),
	}

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

// UPDATED: Now only returns a boolean to confirm OTP was sent
func (h *UserGrpcHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	requiresOTP, err := h.usecase.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return &pb.LoginResponse{RequiresOtp: requiresOTP}, nil
}

// NEW: Verifies OTP and returns the actual tokens
func (h *UserGrpcHandler) VerifyOTP(ctx context.Context, req *pb.VerifyOTPRequest) (*pb.TokenResponse, error) {
	token, err := h.usecase.VerifyOTP(ctx, req.GetEmail(), req.GetOtp(), req.GetUserAgent(), req.GetClientIp())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired OTP")
	}
	return &pb.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		UserId: token.UserId,
		Email: token.Email,
		Role: token.Role,
	}, nil
}

// UPDATED: Now passes UserAgent and ClientIP to Usecase
func (h *UserGrpcHandler) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.TokenResponse, error) {
	token, err := h.usecase.GoogleLogin(ctx, req.GetIdToken(), req.GetUserAgent(), req.GetClientIp())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pb.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		UserId: token.UserId,
		Email: token.Email,
		Role: token.Role,
	}, nil
}

// UPDATED: Now passes UserAgent and ClientIP to Usecase
func (h *UserGrpcHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.TokenResponse, error) {
	token, err := h.usecase.RefreshToken(ctx, req.GetRefreshToken(), req.GetUserAgent(), req.GetClientIp())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	return &pb.TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		UserId: token.UserId,
		Email: token.Email,
		Role: token.Role,
	}, nil
}

func (h *UserGrpcHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	domainUser := &domain.User{
		ID:       req.GetUserId(),
		FullName: req.GetFullName(),
		Phone:    req.GetPhone(),
		Address:  req.GetAddress(),
	}

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

// --- NEW Session Management Handlers ---

func (h *UserGrpcHandler) GetUserSessions(ctx context.Context, req *pb.GetSessionsRequest) (*pb.GetSessionsResponse, error) {
	sessions, err := h.usecase.GetUserSessions(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch sessions")
	}

	var pbSessions []*pb.SessionInfo
	for _, s := range sessions {
		pbSessions = append(pbSessions, &pb.SessionInfo{
			SessionId: s.SessionID,
			UserAgent: s.UserAgent,
			ClientIp:  s.ClientIP,
			CreatedAt: s.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetSessionsResponse{Sessions: pbSessions}, nil
}

func (h *UserGrpcHandler) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*pb.SuccessResponse, error) {
	err := h.usecase.RevokeSession(ctx, req.GetUserId(), req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "session not found")
	}
	return &pb.SuccessResponse{Success: true}, nil
}

func (h *UserGrpcHandler) RevokeAllSessions(ctx context.Context, req *pb.RevokeAllSessionsRequest) (*pb.SuccessResponse, error) {
	err := h.usecase.RevokeAllSessions(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke sessions")
	}
	return &pb.SuccessResponse{Success: true}, nil
}

func (h *UserGrpcHandler) ResendOTP(ctx context.Context, req *pb.ResendOTPRequest) (*pb.SuccessResponse, error) {
	err := h.usecase.ResendOTP(ctx, req.GetEmail())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to resend OTP")
	}
	return &pb.SuccessResponse{Success: true}, nil
}

func (h *UserGrpcHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.SuccessResponse, error) {
	err := h.usecase.Logout(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}
	return &pb.SuccessResponse{Success: true}, nil
}