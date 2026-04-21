package handler

import (
	"net/http"
	"ecommerce/pb"
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
// @Param q query string true "Search term (e.g., 'mouse')"
// @Success 200 {array} map[string]interface{}
// @Router /api/search [get]
func (h *SearchHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	res, err := h.client.SearchProducts(c.Request.Context(), &pb.SearchRequest{Query: query})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res.Products)
}
