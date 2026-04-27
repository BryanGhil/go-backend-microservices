package usecase

import (
	"context"
	"errors"
	"ecommerce/user-service/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userUseCase struct {
	pgRepo    domain.UserRepository
	redisRepo domain.SessionRepository
}

func NewUserUseCase(pg domain.UserRepository, redis domain.SessionRepository) domain.UserUseCase {
	return &userUseCase{
		pgRepo:    pg,
		redisRepo: redis,
	}
}

func (u *userUseCase) Register(ctx context.Context, req *domain.User, password string) (int64, error) {
	// 1. Security Check: Prevent API users from making themselves admins
	if req.Role == "admin" {
		return 0, errors.New("cannot register as admin via public api")
	}

	// 2. Default to buyer if invalid role
	if req.Role != "seller" {
		req.Role = "buyer"
	}

	// 3. Seller Validation: Ensure shop_name is provided
	if req.Role == "seller" {
		if req.SellerProfile == nil || req.SellerProfile.ShopName == "" {
			return 0, errors.New("shop_name is required to register as a seller")
		}
	} else {
		// Ensure buyers don't have dangling profile data
		req.SellerProfile = nil 
	}

	// 4. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	
	// 5. Set Database Defaults
	req.PasswordHash = string(hashedPassword)
	req.IsActive = true
	req.Provider = "local"

	// 6. Save to Database (Repository will handle the transaction for both tables)
	return u.pgRepo.CreateUser(ctx, req)
}

func (u *userUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.pgRepo.GetUserByEmail(ctx, email)
	if err != nil || !user.IsActive {
		return "", errors.New("invalid credentials or inactive account")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	token := uuid.New().String()
	sessionData := &domain.SessionData{
		UserID: user.ID,
		Role:   user.Role,
	}

	err = u.redisRepo.CreateSession(ctx, token, sessionData)
	if err != nil {
		return "", errors.New("failed to create session")
	}

	return token, nil
}

func (u *userUseCase) VerifySession(ctx context.Context, token string) (int64, string, error) {
	sessionData, err := u.redisRepo.GetSessionData(ctx, token)
	if err != nil {
		return 0, "", err
	}
	return sessionData.UserID, sessionData.Role, nil
}

func (u *userUseCase) UpdateProfile(ctx context.Context, req *domain.User) error {
	// Let the repository layer handle updating the users table, 
	// and conditionally updating the seller_profiles table if req.Role == "seller"
	return u.pgRepo.UpdateProfile(ctx, req)
}