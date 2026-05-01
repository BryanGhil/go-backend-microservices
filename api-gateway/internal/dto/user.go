package dto

// Requests
type RegisterReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role"`      // Optional: defaults to "buyer" in usecase
	ShopName string `json:"shop_name"` // Required if role is "seller"
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email" example:"zzz@zzz.com"`
	Password string `json:"password" binding:"required" example:"12345678"`
}

type UpdateProfileReq struct {
	FullName        string `json:"full_name"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	ShopName        string `json:"shop_name"`        // For Sellers
	ShopDescription string `json:"shop_description"` // For Sellers
}

// Optional Standard Swagger Responses (if you don't already have them in a common dto file)
type TokenResponse struct {
	Token string `json:"token"`
}

type UserIDResponse struct {
	UserID int64 `json:"user_id"`
}