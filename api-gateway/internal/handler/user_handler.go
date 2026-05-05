package handler

import (
	"ecommerce/api-gateway/internal/dto"
	"ecommerce/api-gateway/pkg/utils"
	"ecommerce/pb"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	client pb.UserServiceClient
}

func NewUserHandler(client pb.UserServiceClient) *UserHandler {
	return &UserHandler{client: client}
}

func (h *UserHandler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	// Public Auth Routes
	public.POST("/auth/register", h.Register)
	public.POST("/auth/login", h.Login)
	public.POST("/auth/verify-otp", h.VerifyOTP) // NEW: Step 2 of Login
	public.POST("/auth/google", h.GoogleLogin)
	public.POST("/auth/refresh", h.RefreshToken)

	// Protected User Routes
	protected.PUT("/users/profile", h.UpdateProfile)

	// NEW: Protected Session Management Routes
	protected.GET("/users/sessions", h.GetSessions)
	protected.DELETE("/users/sessions/:session_id", h.RevokeSession)
	protected.DELETE("/users/sessions", h.RevokeAllSessions)

	public.POST("/auth/resend-otp", h.ResendOTP)
	public.POST("/auth/logout", h.Logout)
}

// @Summary Register a new user
// @Description Registers a new buyer or seller account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterReq true "User credentials and profile details"
// @Success 201 {object} dto.UserIDResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format or missing fields")
		return
	}

	res, err := h.client.Register(c.Request.Context(), &pb.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     req.Role,
		ShopName: req.ShopName,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, status.Convert(err).Message())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", gin.H{"user_id": res.UserId})
}

// @Summary Update User Profile
// @Description Updates the profile of the currently authenticated user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileReq true "Profile Data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Matches the key set in the AuthMiddleware
	userID := c.GetInt64("userID")

	var req dto.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	_, err := h.client.UpdateProfile(c.Request.Context(), &pb.UpdateProfileRequest{
		UserId:          userID,
		FullName:        req.FullName,
		Phone:           req.Phone,
		Address:         req.Address,
		ShopName:        req.ShopName,
		ShopDescription: req.ShopDescription,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, status.Convert(err).Message())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", nil)
}

// @Summary Step 1: Login user
// @Description Validates credentials and sends a 6-digit OTP to the user's email.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginReq true "User credentials"
// @Success 200 {object} dto.LoginOTPResponse
// @Failure 400 {object} map[string]interface{} "Invalid JSON format"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	_, err := h.client.Login(c.Request.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, status.Convert(err).Message())
		return
	}

	// Login successful, OTP sent. Tell the frontend to show the OTP input screen.
	utils.SuccessResponse(c, http.StatusOK, "OTP sent to email", gin.H{
		"requires_otp": true,
	})
}

// @Summary Step 2: Verify OTP
// @Description Verifies the 6-digit OTP and issues JWT tokens based on client_type.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyOTPReq true "Email, OTP, and client type"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/auth/verify-otp [post]
func (h *UserHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Capture Device Metadata automatically!
	userAgent := c.GetHeader("User-Agent")
	clientIP := c.ClientIP()

	res, err := h.client.VerifyOTP(c.Request.Context(), &pb.VerifyOTPRequest{
		Email:     req.Email,
		Otp:       req.OTP,
		UserAgent: userAgent,
		ClientIp:  clientIP,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, status.Convert(err).Message())
		return
	}

	if req.ClientType == "web" {
		c.SetCookie("refresh_token", res.RefreshToken, 7*24*60*60, "/api/auth", "localhost", false, true)
		utils.SuccessResponse(c, http.StatusOK, "Login successful", dto.TokenResponse{
			AccessToken: res.AccessToken,
			User: dto.TokenUserResponse{
				UserId: res.UserId,
				Email:  res.Email,
				Role:   res.Role,
			},
		})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", dto.TokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		User: dto.TokenUserResponse{
			UserId: res.UserId,
			Email:  res.Email,
			Role:   res.Role,
		},
	})
}

// @Summary Google OAuth Login
// @Description Authenticates via Google ID token. Auto-captures device metadata.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.GoogleLoginReq true "Google ID Token and client type"
// @Success 200 {object} dto.TokenResponse
// @Failure 401 {object} map[string]interface{}
// @Router /api/auth/google [post]
func (h *UserHandler) GoogleLogin(c *gin.Context) {
	var req dto.GoogleLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	res, err := h.client.GoogleLogin(c.Request.Context(), &pb.GoogleLoginRequest{
		IdToken:   req.IDToken,
		UserAgent: c.GetHeader("User-Agent"),
		ClientIp:  c.ClientIP(),
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, status.Convert(err).Message())
		return
	}

	if req.ClientType == "web" {
		c.SetCookie("refresh_token", res.RefreshToken, 7*24*60*60, "/api/auth", "localhost", false, true)
		utils.SuccessResponse(c, http.StatusOK, "Login successful", dto.TokenResponse{
			AccessToken: res.AccessToken,
			User: dto.TokenUserResponse{
				UserId: res.UserId,
				Email:  res.Email,
				Role:   res.Role,
			},
		})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", dto.TokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		User: dto.TokenUserResponse{
			UserId: res.UserId,
			Email:  res.Email,
			Role:   res.Role,
		},
	})
}

// @Summary Refresh Access Token
// @Description Rotates the refresh token and captures new IP if the user changed Wi-Fi.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenReq false "Refresh Token (Required for mobile)"
// @Success 200 {object} dto.TokenResponse
// @Router /api/auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var refreshToken string

	var req dto.RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
        refreshToken = req.RefreshToken
        fmt.Println("[DEBUG] Found token in JSON body")
    }

    // 2. Try Cookie if JSON failed
    if refreshToken == "" {
        cookieToken, err := c.Cookie("refresh_token")
        if err == nil {
            refreshToken = cookieToken
            fmt.Println("[DEBUG] Found token in Cookie ", refreshToken)
        } else {
            fmt.Printf("[DEBUG] Cookie check failed: %v\n", err)
        }
    }

	if refreshToken == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Refresh token is required")
		return
	}

	res, err := h.client.RefreshToken(c.Request.Context(), &pb.RefreshTokenRequest{
		RefreshToken: refreshToken,
		UserAgent:    c.GetHeader("User-Agent"), // Updates the session metadata in Redis!
		ClientIp:     c.ClientIP(),              // Updates IP in Redis!
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	if _, err := c.Cookie("refresh_token"); err == nil {
		c.SetCookie("refresh_token", res.RefreshToken, 7*24*60*60, "/api/auth", "localhost", false, true)
		utils.SuccessResponse(c, http.StatusOK, "Token refreshed", dto.TokenResponse{
			AccessToken: res.AccessToken,
			User: dto.TokenUserResponse{
				UserId: res.UserId,
				Email:  res.Email,
				Role:   res.Role,
			},
		})
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", dto.TokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		User: dto.TokenUserResponse{
			UserId: res.UserId,
			Email:  res.Email,
			Role:   res.Role,
		},
	})
}

// @Summary Get Active Sessions
// @Description Lists all devices currently logged into this account
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.SessionInfoResponse
// @Router /api/users/sessions [get]
func (h *UserHandler) GetSessions(c *gin.Context) {
	userID := c.GetInt64("userID")

	res, err := h.client.GetUserSessions(c.Request.Context(), &pb.GetSessionsRequest{
		UserId: userID,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch sessions")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Sessions retrieved", res.Sessions)
}

// @Summary Revoke Specific Session
// @Description Logs out a specific device
// @Tags Users
// @Security BearerAuth
// @Param session_id path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/users/sessions/{session_id} [delete]
func (h *UserHandler) RevokeSession(c *gin.Context) {
	userID := c.GetInt64("userID")
	sessionID := c.Param("session_id")

	_, err := h.client.RevokeSession(c.Request.Context(), &pb.RevokeSessionRequest{
		UserId:    userID,
		SessionId: sessionID,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Session not found or already revoked")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Session revoked successfully", nil)
}

// @Summary Revoke All Sessions
// @Description Logs the user out of EVERY device
// @Tags Users
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/users/sessions [delete]
func (h *UserHandler) RevokeAllSessions(c *gin.Context) {
	userID := c.GetInt64("userID")

	_, err := h.client.RevokeAllSessions(c.Request.Context(), &pb.RevokeAllSessionsRequest{
		UserId: userID,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to revoke sessions")
		return
	}

	// Also clear the HttpOnly cookie for the current web device
	c.SetCookie("refresh_token", "", -1, "/api/auth", "localhost", true, true)

	utils.SuccessResponse(c, http.StatusOK, "Successfully logged out of all devices", nil)
}

// @Summary Resend OTP
// @Description Sends a new 6-digit OTP to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResendOTPReq true "User Email"
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/resend-otp [post]
func (h *UserHandler) ResendOTP(c *gin.Context) {
	var req dto.ResendOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	_, err := h.client.ResendOTP(c.Request.Context(), &pb.ResendOTPRequest{
		Email: req.Email,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to resend OTP. Please try again.")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "A new OTP has been sent to your email", nil)
}

// @Summary Logout current device
// @Description Logs out the current session and clears the HttpOnly cookie
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LogoutReq false "Refresh Token (Required for mobile)"
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	var refreshToken string

	// 1. Try to get it from the JSON body (Mobile)
	var req dto.LogoutReq
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		refreshToken = req.RefreshToken
	}

	// 2. Try to get it from the Cookie (Web)
	if refreshToken == "" {
		cookieToken, err := c.Cookie("refresh_token")
		if err == nil {
			refreshToken = cookieToken
		}
	}

	if refreshToken == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "No active session found")
		return
	}

	// 3. Tell the User Service to delete the session from Redis
	_, _ = h.client.Logout(c.Request.Context(), &pb.LogoutRequest{
		RefreshToken: refreshToken,
	})

	// 4. Force the browser to delete the HttpOnly cookie by setting Max-Age to -1
	c.SetCookie("refresh_token", "", -1, "/api/auth", "localhost", false, true)

	utils.SuccessResponse(c, http.StatusOK, "Successfully logged out", nil)
}
