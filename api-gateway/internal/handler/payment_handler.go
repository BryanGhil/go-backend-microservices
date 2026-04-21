package handler

import (
	"ecommerce/pb"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	client pb.PaymentServiceClient
}

func NewPaymentHandler(client pb.PaymentServiceClient) *PaymentHandler {
	return &PaymentHandler{client: client}
}

func (h *PaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/payments/:order_id", h.GetStatus)
}

// @Summary Get Payment Status
// @Description Check the raw transaction status from the payment gateway
// @Tags Payments
// @Produce json
// @Param order_id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/payments/{order_id} [get]
func (h *PaymentHandler) GetStatus(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Param("order_id"), 10, 64)
	res, err := h.client.GetPaymentStatus(c.Request.Context(), &pb.GetPaymentStatusRequest{OrderId: orderID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}