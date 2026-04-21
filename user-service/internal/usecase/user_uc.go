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

// THE CONSTRUCTOR: It takes both the Postgres and Redis repositories
func NewUserUseCase(pg domain.UserRepository, redis domain.SessionRepository) domain.UserUseCase {
	return &userUseCase{
		pgRepo:    pg,
		redisRepo: redis,
	}
}

func (u *userUseCase) Register(ctx context.Context, email, password string) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	return u.pgRepo.CreateUser(ctx, email, string(hashedPassword))
}

func (u *userUseCase) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Check Postgres for the user
	user, err := u.pgRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// 2. Compare the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// 3. Generate a UUID token and save to Redis
	token := uuid.New().String()
	err = u.redisRepo.CreateSession(ctx, token, user.ID)
	if err != nil {
		return "", errors.New("failed to create session")
	}

	return token, nil
}

func (u *userUseCase) VerifySession(ctx context.Context, token string) (int64, error) {
	// Just ask Redis if the token exists
	return u.redisRepo.GetUserIDByToken(ctx, token)
}

func (u *userUseCase) UpdateProfile(ctx context.Context, id int64, name, phone, address string) error {
	// Save to Postgres
	return u.pgRepo.UpdateProfile(ctx, id, name, phone, address)
}