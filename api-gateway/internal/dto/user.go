package dto

// --- Requests ---

type RegisterReq struct {
	Email    string `json:"email" binding:"required,email" example:"zzz@zzz.com"`
	Password string `json:"password" binding:"required,min=6" example:"12345678"`
	FullName string `json:"full_name" binding:"required" example:"Zzz Zzz"`
	Role     string `json:"role" example:"seller"`       // Optional: defaults to "buyer" in usecase
	ShopName string `json:"shop_name" example:"Zz Shop"` // Required if role is "seller"
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email" example:"zzz@zzz.com"`
	Password string `json:"password" binding:"required" example:"12345678"`
}

type VerifyOTPReq struct {
	Email      string `json:"email" binding:"required,email" example:"zzz@zzz.com"`
	OTP        string `json:"otp" binding:"required,len=6" example:"123456"`
	ClientType string `json:"client_type" binding:"required,oneof=web mobile" example:"web"`
}

type GoogleLoginReq struct {
	IDToken    string `json:"id_token" binding:"required" example:"eyJhbGciOiJSUzI1NiIs..."`
	ClientType string `json:"client_type" binding:"required,oneof=web mobile" example:"mobile"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" example:"your-uuid-refresh-token"`
}

type UpdateProfileReq struct {
	FullName        string `json:"full_name" example:"Zzz"`
	Phone           string `json:"phone" example:"0812345678"`
	Address         string `json:"address" example:"ZZ Street"`
	ShopName        string `json:"shop_name" example:"ZZ Shop"`                    // For Sellers
	ShopDescription string `json:"shop_description" example:"Welcome to ZZZ Shop"` // For Sellers
}

// --- Responses ---

type TokenResponse struct {
	AccessToken  string            `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string            `json:"refresh_token,omitempty" example:"your-uuid-refresh-token"`
	User         TokenUserResponse `json:"user" `
}

type TokenUserResponse struct {
	UserId int64  `json:"user_id" example:"1"`
	Email  string `json:"email" example:"example@gmail.com"`
	Role   string `json:"role" example:"seller"`
}

type UserIDResponse struct {
	UserID int64 `json:"user_id"`
}

type LoginOTPResponse struct {
	Message     string `json:"message" example:"OTP sent to email"`
	RequiresOTP bool   `json:"requires_otp" example:"true"`
}

type SessionInfoResponse struct {
	SessionID string `json:"session_id" example:"uuid-string"`
	UserAgent string `json:"user_agent" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64)..."`
	ClientIP  string `json:"client_ip" example:"192.168.1.5"`
	CreatedAt string `json:"created_at" example:"2024-05-12T15:04:05Z"`
}

// dto/user.go (Add these to your Request section)

type ResendOTPReq struct {
	Email string `json:"email" binding:"required,email" example:"zzz@zzz.com"`
}

type LogoutReq struct {
	// Not required because web clients send it via cookie
	RefreshToken string `json:"refresh_token" example:"your-uuid-refresh-token"`
}
