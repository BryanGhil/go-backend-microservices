package handler

import (
	"ecommerce/api-gateway/internal/dto"
	"ecommerce/api-gateway/pkg/utils"
	"ecommerce/pb"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

type InventoryHandler struct {
	client pb.InventoryServiceClient
}

func NewInventoryHandler(client pb.InventoryServiceClient) *InventoryHandler {
	return &InventoryHandler{client: client}
}

func (h *InventoryHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/inventory/:id", h.GetStock)
	// Note: You should likely place this under an Admin/Seller protected route group in main.go
	router.POST("/inventory/adjust", h.AdjustStock) 
}

// @Summary Get Product Stock
// @Tags Inventory
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/inventory/{id} [get]
func (h *InventoryHandler) GetStock(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	
	res, err := h.client.GetStock(c.Request.Context(), &pb.GetStockRequest{ProductId: id})
	if err != nil {
		// Use status.Convert to extract the clean message from gRPC
		utils.ErrorResponse(c, http.StatusInternalServerError, status.Convert(err).Message())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Stock retrieved successfully", gin.H{
		"product_id": id,
		"stock":      res.Stock,
	})
}

// @Summary Add Physical Stock
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body dto.AddStockReq true "Stock Details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/inventory/add [post]
func (h *InventoryHandler) AdjustStock(c *gin.Context) {
	var req dto.AddStockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format or missing fields")
		return
	}

	_, err := h.client.AdjustStock(c.Request.Context(), &pb.AdjustStockRequest{
		ProductId: req.ProductID,
		Quantity:  req.Quantity,
	})
	
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, status.Convert(err).Message())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Stock added successfully", gin.H{
		"product_id":     req.ProductID,
		"added_quantity": req.Quantity,
	})
}