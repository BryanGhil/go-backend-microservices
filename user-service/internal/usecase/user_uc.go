package usecase

import (
	"context"
	"errors"
	"time"

	"ecommerce/user-service/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// In a real production app, inject this from your environment variables (.env)
var jwtSecretKey = []byte("your-super-secret-key-change-me")

// Define the structure of your JWT payload
type jwtCustomClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

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
	if req.Role == "admin" {
		return 0, errors.New("cannot register as admin via public api")
	}

	if req.Role != "seller" {
		req.Role = "buyer"
	}

	if req.Role == "seller" {
		if req.SellerProfile == nil || req.SellerProfile.ShopName == "" {
			return 0, errors.New("shop_name is required to register as a seller")
		}
	} else {
		req.SellerProfile = nil 
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	
	req.PasswordHash = string(hashedPassword)
	req.IsActive = true
	req.Provider = "local"

	return u.pgRepo.CreateUser(ctx, req)
}

func (u *userUseCase) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Check Postgres for the user
	user, err := u.pgRepo.GetUserByEmail(ctx, email)
	if err != nil || !user.IsActive {
		return "", errors.New("invalid credentials or inactive account")
	}

	// 2. Compare the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// 3. Create the JWT Claims
	claims := &jwtCustomClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hour expiration
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(), // Unique ID for this specific token
		},
	}

	// 4. Generate and Sign the JWT
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := tokenObj.SignedString(jwtSecretKey)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	// 5. Save session in Redis 
	// We use the full JWT string as the key to maintain compatibility with your VerifySession logic
	sessionData := &domain.SessionData{
		UserID: user.ID,
		Role:   user.Role,
	}

	err = u.redisRepo.CreateSession(ctx, tokenString, sessionData)
	if err != nil {
		return "", errors.New("failed to create session")
	}

	return tokenString, nil
}

func (u *userUseCase) VerifySession(ctx context.Context, token string) (int64, string, error) {
	// Right now, this just asks Redis if the token exists.
	sessionData, err := u.redisRepo.GetSessionData(ctx, token)
	if err != nil {
		return 0, "", err
	}
	return sessionData.UserID, sessionData.Role, nil
}

func (u *userUseCase) UpdateProfile(ctx context.Context, req *domain.User) error {
	return u.pgRepo.UpdateProfile(ctx, req)
}