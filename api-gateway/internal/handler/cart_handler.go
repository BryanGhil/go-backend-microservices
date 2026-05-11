package handler

import (
	"net/http"
	"strconv"

	"ecommerce/api-gateway/internal/dto"
	"ecommerce/api-gateway/pkg/utils"
	"ecommerce/pb"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

type CartHandler struct {
	client pb.CartServiceClient
}

func NewCartHandler(client pb.CartServiceClient) *CartHandler {
	return &CartHandler{client: client}
}

func (h *CartHandler) RegisterRoutes(router *gin.RouterGroup) {
	// These routes MUST be protected by your JWT Auth Middleware!
	cartRoutes := router.Group("/cart")
	{
		cartRoutes.GET("/", h.GetCart)
		cartRoutes.POST("/add", h.AddToCart)
		cartRoutes.DELETE("/remove/:product_id", h.RemoveItem)
		cartRoutes.DELETE("/clear", h.ClearCart)
		cartRoutes.GET("/count", h.GetCartCount)
	}
}

// @Summary Get Current User's Cart
// @Description Fetches the complete cart with real-time product prices and details
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/cart/ [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	// Extract user ID from the JWT auth middleware
	userID := c.GetInt64("userID")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	res, err := h.client.GetCart(c.Request.Context(), &pb.GetCartRequest{UserId: userID})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, status.Convert(err).Message())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cart retrieved successfully", res.Items)
}

// @Summary Add or Update Cart Item
// @Description Adds a product to the cart or updates its quantity. Send negative quantity to subtract.
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AddToCartReq true "Product and Quantity"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/cart/add [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.AddToCartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format or missing fields")
		return
	}

	_, err := h.client.AddToCart(c.Request.Context(), &pb.AddToCartRequest{
		UserId:    userID,
		ProductId: req.ProductID,
		Quantity:  req.Quantity,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, status.Convert(err).Message())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Item added to cart successfully", nil)
}

// @Summary Remove Item from Cart
// @Description Completely deletes a product from the user's cart (like clicking the trash icon)
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param product_id path int true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/cart/remove/{product_id} [delete]
func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	productID, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	_, err = h.client.RemoveItem(c.Request.Context(), &pb.RemoveItemRequest{
		UserId:    userID,
		ProductId: productID,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, status.Convert(err).Message())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Item removed from cart", nil)
}

// @Summary Clear Entire Cart
// @Description Empties the cart. Usually called automatically after successful payment.
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/cart/clear [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_, err := h.client.ClearCart(c.Request.Context(), &pb.ClearCartRequest{UserId: userID})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, status.Convert(err).Message())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cart cleared successfully", nil)
}

// @Summary Get Cart Badge Count
// @Description Super fast endpoint to get total number of items for the frontend cart icon badge
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/cart/count [get]
func (h *CartHandler) GetCartCount(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	res, err := h.client.GetCartCount(c.Request.Context(), &pb.GetCartCountRequest{UserId: userID})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, status.Convert(err).Message())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cart count retrieved", gin.H{
		"count": res.Count,
	})
}