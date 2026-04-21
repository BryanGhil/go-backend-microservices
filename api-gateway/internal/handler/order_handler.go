package handler

import (
	"ecommerce/pb"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	client pb.OrderServiceClient
}

func NewOrderHandler(client pb.OrderServiceClient) *OrderHandler {
	return &OrderHandler{client: client}
}

func (h *OrderHandler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.POST("/checkout", h.CreateOrder)
	protected.GET("/orders/:id", h.GetStatus)
}

type CheckoutReq struct {
	ProductID int64   `json:"product_id" example:"1"`
	Amount    float64 `json:"amount" example:"49.99"`
}

// @Summary Create an Order (Checkout)
// @Description Initiates the Saga pattern for ordering a product
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CheckoutReq true "Checkout Details"
// @Success 202 {object} map[string]interface{}
// @Router /api/checkout [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.GetInt64("userID") // From Auth Middleware
	var req CheckoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	res, err := h.client.CreateOrder(c.Request.Context(), &pb.CreateOrderRequest{
		UserId:    userID,
		ProductId: req.ProductID,
		Amount:    float32(req.Amount),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Saga has started in the background!
	c.JSON(http.StatusAccepted, gin.H{
		"order_id": res.OrderId,
		"message":  "Order is being processed. Please poll the status endpoint.",
	})
}

// @Summary Get Order Status
// @Description Check if an order is PENDING, COMPLETED, or CANCELLED
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/{id} [get]
func (h *OrderHandler) GetStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	res, err := h.client.GetOrderStatus(c.Request.Context(), &pb.GetOrderStatusRequest{OrderId: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}