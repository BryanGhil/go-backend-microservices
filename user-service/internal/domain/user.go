package domain

import (
	"context"
	"time"
)

// Matches the "users" table
type User struct {
	ID           int64
	Email        string
	PasswordHash string
	FullName     string
	Phone        string
	Address      string
	Role         string
	IsActive     bool
	LastLoginAt  *time.Time // Pointer because it can be NULL
	Provider     string
	ProviderID   string

	// The Relationship: Nil if the user is a "buyer" or "admin"
	SellerProfile *SellerProfile 
}

// Matches the "seller_profiles" table
type SellerProfile struct {
	UserID          int64
	ShopName        string
	ShopDescription string
	TaxID           string
	IsVerified      bool
	Rating          float64
}

// Struct to hold data retrieved from Redis
type SessionData struct {
	UserID int64
	Role   string
}

// 1. Postgres handles permanent data
type UserRepository interface {
	// The repo will need to insert into BOTH tables using a DB Transaction
	CreateUser(ctx context.Context, user *User) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateProfile(ctx context.Context, user *User) error
}

// 2. Redis handles temporary sessions
type SessionRepository interface {
	CreateSession(ctx context.Context, token string, data *SessionData) error
	GetSessionData(ctx context.Context, token string) (*SessionData, error)
}

// 3. The UseCase handles the business rules
type UserUseCase interface {
	Register(ctx context.Context, req *User, password string) (int64, error)
	Login(ctx context.Context, email, password string) (string, error)
	VerifySession(ctx context.Context, token string) (int64, string, error) 
	UpdateProfile(ctx context.Context, req *User) error
}