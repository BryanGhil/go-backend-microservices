package handler

import (
	"net/http"
	"ecommerce/pb"
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

// --- DTOs (Data Transfer Objects) for Swagger ---
type RegisterReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileReq struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

// @Summary Register a new user
// @Tags Users
// @Accept json
// @Produce json
// @Param request body RegisterReq true "User credentials"
// @Success 201 {object} map[string]interface{}
// @Router /api/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	res, err := h.client.Register(c.Request.Context(), &pb.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": status.Convert(err).Message()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user_id": res.UserId})
}

// @Summary Login user
// @Tags Users
// @Accept json
// @Produce json
// @Param request body LoginReq true "User credentials"
// @Success 200 {object} map[string]interface{}
// @Router /api/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	res, err := h.client.Login(c.Request.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": status.Convert(err).Message()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": res.Token})
}

// @Summary Update User Profile
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateProfileReq true "Profile Data"
// @Success 200 {object} map[string]interface{}
// @Router /api/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetInt64("userID")
	var req UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	_, err := h.client.UpdateProfile(c.Request.Context(), &pb.UpdateProfileRequest{
		UserId:   userID,
		FullName: req.FullName,
		Phone:    req.Phone,
		Address:  req.Address,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": status.Convert(err).Message()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}