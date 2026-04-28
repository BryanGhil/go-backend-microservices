package handler

import (
	"ecommerce/api-gateway/pkg/utils"
	"ecommerce/pb"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	client pb.SearchServiceClient
}

func NewSearchHandler(client pb.SearchServiceClient) *SearchHandler {
	return &SearchHandler{client: client}
}

func (h *SearchHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/search", h.SearchProducts)
}

// @Summary Search products
// @Description Query the Elasticsearch cluster for products matching the search term
// @Tags Search
// @Produce json
// @Param q query string true "Search term (e.g., 'keyboard')"
// @Param limit query int false "Items per page (default 10)"
// @Param page query int false "Page number (default 1)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/search [get]
func (h *SearchHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Search query 'q' is required")
		return
	}

	// 1. Extract Pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	// 2. Pass pagination to the gRPC client
	res, err := h.client.SearchProducts(c.Request.Context(), &pb.SearchRequest{
		Query:  query,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search products")
		return
	}

	// 3. Format Response nicely to hide ugly Protobuf generated fields
	var products []map[string]interface{}
	for _, p := range res.Products {
		products = append(products, map[string]interface{}{
			"id":          p.Id,
			"seller_id":   p.SellerId,
			"name":        p.Name,
			"description": p.Description,
			"category":    p.Category,
			"price":       p.Price,
			"image_url":   p.ImageUrl,
		})
	}

	// 4. Return the standardized response including the Total Count
	utils.SuccessResponse(c, http.StatusOK, "Search successful", gin.H{
		"products":    products,
		"total_count": res.TotalCount,
		"page":        page,
		"limit":       limit,
	})
}