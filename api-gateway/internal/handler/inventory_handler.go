package handler

import (
	"ecommerce/pb"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	client pb.InventoryServiceClient
}

func NewInventoryHandler(client pb.InventoryServiceClient) *InventoryHandler {
	return &InventoryHandler{client: client}
}

func (h *InventoryHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/inventory/:id", h.GetStock)
	router.POST("/inventory/add", h.AddStock)
}

type AddStockReq struct {
	ProductID int64 `json:"product_id" example:"1"`
	Quantity  int32 `json:"quantity" example:"50"`
}

// @Summary Get Product Stock
// @Tags Inventory
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/inventory/{id} [get]
func (h *InventoryHandler) GetStock(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	res, err := h.client.GetStock(c.Request.Context(), &pb.GetStockRequest{ProductId: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// @Summary Add Physical Stock
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body AddStockReq true "Stock Details"
// @Success 200 {object} map[string]interface{}
// @Router /api/inventory/add [post]
func (h *InventoryHandler) AddStock(c *gin.Context) {
	var req AddStockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	res, err := h.client.AddStock(c.Request.Context(), &pb.AddStockRequest{
		ProductId: req.ProductID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}