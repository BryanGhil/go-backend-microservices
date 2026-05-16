package handler

import (
	"ecommerce/pb"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	client pb.PaymentServiceClient
}

func NewPaymentHandler(client pb.PaymentServiceClient) *PaymentHandler {
	return &PaymentHandler{client: client}
}

func (h *PaymentHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/payments/:correlation_id", h.GetStatus)
}

// @Summary Get Payment Status
// @Description Check the raw transaction status from the payment gateway
// @Tags Payments
// @Produce json
// @Param correlation_id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/payments/{correlation_id} [get]
func (h *PaymentHandler) GetStatus(c *gin.Context) {
	correlationID := c.Param("correlation_id")
	res, err := h.client.GetPaymentStatus(c.Request.Context(), &pb.GetPaymentStatusRequest{CorrelationId: correlationID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}