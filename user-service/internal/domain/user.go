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
	SessionID string    `json:"session_id"` // A safe UUID just for identifying the session in the UI
	UserID    int64     `json:"user_id"`
	Role      string    `json:"role"`
	UserAgent string    `json:"user_agent"` // e.g., "Chrome on Mac" or "iPhone 15"
	ClientIP  string    `json:"client_ip"`  // e.g., "192.168.1.5"
	CreatedAt time.Time `json:"created_at"`
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
	StoreRefreshToken(ctx context.Context, token string, data *SessionData) error
	GetSessionData(ctx context.Context, token string) (*SessionData, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	GetUserSessions(ctx context.Context, userID int64) ([]*SessionData, error)

	StoreOTP(ctx context.Context, email string, otp string) error
	GetOTP(ctx context.Context, email string) (string, error)
	DeleteOTP(ctx context.Context, email string) error

	RevokeSession(ctx context.Context, userID int64, sessionID string) error
	RevokeAllSessions(ctx context.Context, userID int64) error
}

// 3. The UseCase handles the business rules
type UserUseCase interface {
	Register(ctx context.Context, req *User, password string) (int64, error)

	Login(ctx context.Context, email, password string) (bool, error)
	GoogleLogin(ctx context.Context, googleIDToken, userAgent, clientIP string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken, userAgent, clientIP string) (string, string, error)

	// NEW: Verify OTP actually returns the tokens
	VerifyOTP(ctx context.Context, email, otp, userAgent, clientIP string) (string, string, error)

	GetUserSessions(ctx context.Context, userID int64) ([]*SessionData, error)

	RevokeSession(ctx context.Context, userID int64, sessionID string) error
	RevokeAllSessions(ctx context.Context, userID int64) error

	UpdateProfile(ctx context.Context, req *User) error

	ResendOTP(ctx context.Context, email string) error
	Logout(ctx context.Context, refreshToken string) error
}
