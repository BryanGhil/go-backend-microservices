package handler

import (
	"ecommerce/api-gateway/pkg/utils"
	"ecommerce/pb"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductHandler struct {
	client pb.ProductServiceClient
}

func NewProductHandler(client pb.ProductServiceClient) *ProductHandler {
	return &ProductHandler{client: client}
}

func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/products", h.CreateProduct)
	router.GET("/products", h.ListProducts)
	router.GET("/products/:id", h.GetProduct)
	router.PUT("/products/:id", h.UpdateProduct)
	router.DELETE("/products/:id", h.DeleteProduct)
}

// --- DTOs for Swagger ---

type ProductReq struct {
	Name  string  `json:"name" binding:"required" example:"Wireless Mouse"`
	Price float64 `json:"price" binding:"required" example:"49.99"`
}

// @Summary Create a new product
// @Description Adds a new product to the catalog and publishes to Kafka
// @Tags Products
// @Accept json
// @Produce json
// @Param request body ProductReq true "Product Details"
// @Success 201 {object} map[string]interface{}
// @Router /api/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var reqBody ProductReq
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}
	res, err := h.client.CreateProduct(c.Request.Context(), &pb.CreateProductRequest{
		Name:  reqBody.Name,
		Price: float32(reqBody.Price),
	})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create product")
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Product created successfully", gin.H{"id": res.Id})
}

// @Summary List all products
// @Description Retrieves all products from the catalog
// @Tags Products
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /api/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	res, err := h.client.ListProducts(c.Request.Context(), &pb.ListProductsRequest{})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch products")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Products retrieved successfully", res.Products)
}

// @Summary Get a product by ID
// @Description Fetches a single product's details
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	res, err := h.client.GetProduct(c.Request.Context(), &pb.GetProductRequest{Id: id})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			utils.ErrorResponse(c, http.StatusNotFound, st.Message())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Product found", res)
}

// @Summary Update a product
// @Description Updates an existing product's name and price
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body ProductReq true "Updated Product Details"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var reqBody ProductReq
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}
	_, err := h.client.UpdateProduct(c.Request.Context(), &pb.UpdateProductRequest{
		Id:    id,
		Name:  reqBody.Name,
		Price: float32(reqBody.Price),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			utils.ErrorResponse(c, http.StatusNotFound, st.Message())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Product updated successfully", gin.H{"id": id})
}

// @Summary Delete a product
// @Description Removes a product from the catalog
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	_, err := h.client.DeleteProduct(c.Request.Context(), &pb.DeleteProductRequest{Id: id})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Product deleted successfully", gin.H{"id": id})
}