package handler

import (
	"ecommerce/api-gateway/internal/dto"
	"ecommerce/api-gateway/pkg/utils"
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
	protected.GET("/orders", h.GetUserOrders)
}

// @Summary Create an Order (Checkout)
// @Description Initiates the Saga pattern for ordering one or multiple products
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CheckoutReq true "product_id and quantity"
// @Success 202 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/checkout [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// Extract the user ID from Auth Middleware
	userID := c.GetInt64("userID")
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.CheckoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format or missing required fields")
		return
	}

	if len(req.Items) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Checkout items cannot be empty")
		return
	}

	// Map the Frontend JSON DTOs to the gRPC Protobuf Messages
	var pbItems []*pb.CheckoutItem
	for _, item := range req.Items {
		pbItems = append(pbItems, &pb.CheckoutItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	res, err := h.client.CreateOrder(c.Request.Context(), &pb.CreateOrderRequest{
		UserId: userID,
		Items:  pbItems,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to initiate checkout process")
		return
	}

	// Saga has started! Return the Correlation ID so the frontend can route the user to payment.
	utils.SuccessResponse(c, http.StatusAccepted, "Order is being processed.", gin.H{
		"correlation_id": res.CorrelationId,
	})
}

// @Summary Get Order Status
// @Description Check if an individual order is PENDING, COMPLETED, or CANCELLED
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/{id} [get]
func (h *OrderHandler) GetStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID format")
		return
	}

	res, err := h.client.GetOrderStatus(c.Request.Context(), &pb.GetOrderStatusRequest{OrderId: id})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch order status")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Status retrieved successfully", gin.H{
		"order_id": id,
		"status":   res.Status,
	})
}

// @Summary Get User Orders
// @Description Fetch all orders for the authenticated user
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/orders [get]
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID := c.GetInt64("userID") // From Auth Middleware

	res, err := h.client.GetUserOrders(c.Request.Context(), &pb.GetUserOrdersRequest{
		UserId: userID,
	})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Orders retrieved successfully", res.Orders)
}
