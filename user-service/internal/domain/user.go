package domain

import "context"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	Address      string
}

// 1. Postgres handles permanent data
type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateProfile(ctx context.Context, id int64, name, phone, address string) error
}

// 2. Redis handles temporary sessions
type SessionRepository interface {
	CreateSession(ctx context.Context, token string, userID int64) error
	GetUserIDByToken(ctx context.Context, token string) (int64, error)
}

// 3. The UseCase handles the business rules
type UserUseCase interface {
	Register(ctx context.Context, email, password string) (int64, error)
	Login(ctx context.Context, email, password string) (string, error)
	VerifySession(ctx context.Context, token string) (int64, error)
	UpdateProfile(ctx context.Context, id int64, name, phone, address string) error
}