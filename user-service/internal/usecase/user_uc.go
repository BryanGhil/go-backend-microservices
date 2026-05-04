package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"ecommerce/user-service/internal/domain"
	"ecommerce/user-service/pkg/sender"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

// In a real production app, inject this from your environment variables (.env)
var jwtSecretKey = []byte(os.Getenv("JWT_SECRET"))
var googleClientID = os.Getenv("GOOGLE_CLIENT_ID") // From Google Cloud Console

// Define the structure of your JWT payload
type jwtCustomClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type userUseCase struct {
	pgRepo         domain.UserRepository
	redisRepo      domain.SessionRepository
	jwtSecret      []byte
	googleClientID string
	emailSender    email.Sender // <--- ADD THIS
}

func NewUserUseCase(pg domain.UserRepository, redis domain.SessionRepository, jwtSecret string, googleClientID string, sender email.Sender) domain.UserUseCase {
	return &userUseCase{
		pgRepo:         pg,
		redisRepo:      redis,
		jwtSecret:      []byte(jwtSecret),
		googleClientID: googleClientID,
		emailSender:    sender, // <--- ADD THIS
	}
}

func generateRandomOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (u *userUseCase) generateTokenPair(ctx context.Context, user *domain.User, userAgent string, clientIP string) (string, string, error) {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.generateTokenPair")
	defer span.End()

	accessClaims := &jwtCustomClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(jwtSecretKey)

	refreshToken := uuid.New().String()

	sessionData := &domain.SessionData{
		SessionID: uuid.New().String(),
		UserID:    user.ID,
		Role:      user.Role,
		UserAgent: userAgent,
		ClientIP:  clientIP,
		CreatedAt: time.Now(),
	}
	
	err := u.redisRepo.StoreRefreshToken(ctx, refreshToken, sessionData)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store refresh session")
		return "", "", errors.New("failed to store refresh session")
	}

	return accessToken, refreshToken, nil
}

func (u *userUseCase) Register(ctx context.Context, req *domain.User, password string) (int64, error) {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.Register")
	defer span.End()

	if req.Role == "admin" {
		err := errors.New("cannot register as admin via public api")
		span.RecordError(err)
		return 0, err
	}

	if req.Role != "seller" {
		req.Role = "buyer"
	}

	if req.Role == "seller" {
		if req.SellerProfile == nil || req.SellerProfile.ShopName == "" {
			err := errors.New("shop_name is required to register as a seller")
			span.RecordError(err)
			return 0, err
		}
	} else {
		req.SellerProfile = nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to hash password")
		return 0, err
	}

	req.PasswordHash = string(hashedPassword)
	req.IsActive = true
	req.Provider = "local"

	return u.pgRepo.CreateUser(ctx, req)
}

func (u *userUseCase) Login(ctx context.Context, email, password string) (bool, error) {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.Login")
	defer span.End()

	user, err := u.pgRepo.GetUserByEmail(ctx, email)
	if err != nil || !user.IsActive {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid credentials or inactive account")
		return false, errors.New("invalid credentials")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "password mismatch")
		return false, errors.New("invalid credentials")
	}

	otp := generateRandomOTP()
	err = u.redisRepo.StoreOTP(ctx, email, otp)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store OTP")
		return false, errors.New("failed to generate OTP")
	}

	fmt.Println("OTP = " + otp)

	go func() {
		err := u.emailSender.SendOTP(email, otp)
		if err != nil {
			fmt.Printf("[ERROR] Failed to send OTP to %s: %v\n", email, err)
		}
	}()

	return true, nil
}

func (u *userUseCase) VerifyOTP(ctx context.Context, email, otp, userAgent, clientIP string) (string, string, error) {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.VerifyOTP")
	defer span.End()

	storedOTP, err := u.redisRepo.GetOTP(ctx, email)
	if err != nil || storedOTP != otp {
		err := errors.New("invalid or expired OTP")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", "", err
	}

	user, err := u.pgRepo.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user not found after OTP verification")
		return "", "", errors.New("user not found")
	}

	u.redisRepo.DeleteOTP(ctx, email)

	return u.generateTokenPair(ctx, user, userAgent, clientIP)
}

func (u *userUseCase) GoogleLogin(ctx context.Context, googleIDToken string, userAgent string, clientIP string) (string, string, error) {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.GoogleLogin")
	defer span.End()

	// 1. Verify token with Google
	payload, err := idtoken.Validate(ctx, googleIDToken, googleClientID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid google token")
		return "", "", errors.New("invalid google token")
	}

	email := payload.Claims["email"].(string)
	name := payload.Claims["name"].(string)

	// 2. Check if user exists
	user, err := u.pgRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// User doesn't exist, Auto-Register them!
		newUser := &domain.User{
			Email:        email,
			FullName:     name,
			Role:         "buyer", // Default role
			IsActive:     true,
			Provider:     "google",
			ProviderID:   payload.Subject,
			PasswordHash: "", // No password for Google users
		}
		
		id, err := u.pgRepo.CreateUser(ctx, newUser)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to auto-register google user")
			return "", "", errors.New("failed to create google user")
		}
		newUser.ID = id
		user = newUser
	}

	// 3. Return tokens
	return u.generateTokenPair(ctx, user, userAgent, clientIP)
}

func (u *userUseCase) RefreshToken(ctx context.Context, refreshToken string, userAgent string, clientIP string) (string, string, error) {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.RefreshToken")
	defer span.End()

	// 1. Validate refresh token in Redis
	session, err := u.redisRepo.GetSessionData(ctx, refreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid or expired refresh token")
		return "", "", errors.New("invalid or expired refresh token")
	}

	// 2. Create a dummy user object to pass to token generator
	user := &domain.User{
		ID:   session.UserID,
		Role: session.Role,
	}

	// 3. Delete old refresh token (Token Rotation for security)
	u.redisRepo.DeleteRefreshToken(ctx, refreshToken)

	// 4. Generate new pair
	return u.generateTokenPair(ctx, user, userAgent, clientIP)
}

func (u *userUseCase) UpdateProfile(ctx context.Context, req *domain.User) error {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.UpdateProfile")
	defer span.End()

	err := u.pgRepo.UpdateProfile(ctx, req)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (u *userUseCase) GetUserSessions(ctx context.Context, userID int64) ([]*domain.SessionData, error) {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.GetUserSessions")
	defer span.End()

	sessions, err := u.redisRepo.GetUserSessions(ctx, userID)
	if err != nil {
		span.RecordError(err)
	}
	return sessions, err
}

// NEW: Expose Revocation to gRPC
func (u *userUseCase) RevokeSession(ctx context.Context, userID int64, sessionID string) error {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.RevokeSession")
	defer span.End()

	err := u.redisRepo.RevokeSession(ctx, userID, sessionID)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (u *userUseCase) RevokeAllSessions(ctx context.Context, userID int64) error {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.RevokeAllSessions")
	defer span.End()

	err := u.redisRepo.RevokeAllSessions(ctx, userID)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

// usecase/user_uc.go

func (u *userUseCase) ResendOTP(ctx context.Context, email string) error {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.ResendOTP")
	defer span.End()

	// 1. Verify the user actually exists
	user, err := u.pgRepo.GetUserByEmail(ctx, email)
	if err != nil || !user.IsActive {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user not found or inactive")
		return errors.New("invalid email")
	}

	// 2. Generate and store new OTP
	otp := generateRandomOTP()
	err = u.redisRepo.StoreOTP(ctx, email, otp)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store OTP")
		return errors.New("failed to generate OTP")
	}

	// 3. Send the email asynchronously
	go func() {
		err := u.emailSender.SendOTP(email, otp)
		if err != nil {
			fmt.Printf("[ERROR] Failed to resend OTP to %s: %v\n", email, err)
		}
	}()

	return nil
}

func (u *userUseCase) Logout(ctx context.Context, refreshToken string) error {
	tracer := otel.Tracer("user-usecase")
	ctx, span := tracer.Start(ctx, "UseCase.Logout")
	defer span.End()

	// Delete the token from Redis. If it's already gone, we don't care, 
	// the user is effectively logged out anyway!
	err := u.redisRepo.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		span.RecordError(err)
	}
	
	return nil
}