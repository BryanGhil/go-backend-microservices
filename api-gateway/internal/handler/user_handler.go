package handler

import (
	"ecommerce/api-gateway/internal/dto"
	"ecommerce/api-gateway/pkg/utils"
	"ecommerce/pb"
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
	public.POST("/register", h.Register)
	public.POST("/login", h.Login)
	protected.PUT("/users/profile", h.UpdateProfile)
}

// @Summary Register a new user
// @Description Registers a new buyer or seller account
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dto.RegisterReq true "User credentials and profile details"
// @Success 201 {object} dto.UserIDResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/register [post]
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

// @Summary Login user
// @Description Authenticates a user and returns a JWT/Session token
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dto.LoginReq true "User credentials"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	res, err := h.client.Login(c.Request.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, status.Convert(err).Message())
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Login successful", gin.H{"token": res.Token})
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