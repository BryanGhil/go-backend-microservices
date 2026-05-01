package handler

import (
	"ecommerce/api-gateway/internal/dto"
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
	// Assuming you have an AuthMiddleware that sets "user_id" and "role" in the context
	router.POST("/products", h.CreateProduct)
	router.GET("/products", h.ListProducts)
	router.GET("/products/:id", h.GetProduct)
	router.PUT("/products/:id", h.UpdateProduct)
	router.DELETE("/products/:id", h.DeleteProduct)
}

// @Summary Create a new product
// @Description Adds a new product to the catalog. Only Sellers and Admins can do this.
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ProductReq true "Product Details"
// @Success 201 {object} dto.IDResponse
// @Router /api/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// 1. RBAC Check: Get ID and Role from Auth Middleware
	role := c.GetString("role")
	if role != "seller" && role != "admin" {
		utils.ErrorResponse(c, http.StatusForbidden, "Only sellers can create products")
		return
	}
	userID := c.GetInt64("userID")

	var reqBody dto.ProductReq
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// 2. Pass the new fields and the JWT seller_id to gRPC
	res, err := h.client.CreateProduct(c.Request.Context(), &pb.CreateProductRequest{
		SellerId:    userID,
		Name:        reqBody.Name,
		Description: reqBody.Description,
		Category:    reqBody.Category,
		Price:       float64(reqBody.Price),
		ImageUrl:    reqBody.ImageURL,
	})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create product")
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Product created successfully", gin.H{"id": res.Id})
}

// @Summary List all products
// @Description Retrieves all products with pagination and optional filters
// @Tags Products
// @Produce json
// @Param limit query int false "Items per page"
// @Param page query int false "Page number"
// @Param category query string false "Filter by category"
// @Param seller_id query int false "Filter by seller"
// @Success 200 {object} dto.ProductListResponse
// @Router /api/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// 1. Extract Pagination from Query Params
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit // Convert page to SQL offset

	// 2. Extract Filters
	category := c.Query("category")
	sellerID, _ := strconv.ParseInt(c.Query("seller_id"), 10, 64)

	res, err := h.client.ListProducts(c.Request.Context(), &pb.ListProductsRequest{
		Limit:    int32(limit),
		Offset:   int32(offset),
		Category: category,
		SellerId: sellerID,
	})
	
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch products")
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

	utils.SuccessResponse(c, http.StatusOK, "Products retrieved successfully", gin.H{
		"products":    products,
		"total_count": res.TotalCount,
		"page":        page,
		"limit":       limit,
	})
}

// @Summary Get a product by ID
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
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

	// Unpack the nested res.Product
	p := res.Product
	productData := map[string]interface{}{
		"id":          p.Id,
		"seller_id":   p.SellerId,
		"name":        p.Name,
		"description": p.Description,
		"category":    p.Category,
		"price":       p.Price,
		"image_url":   p.ImageUrl,
	}

	utils.SuccessResponse(c, http.StatusOK, "Product found", productData)
}

// @Summary Update a product
// @Tags Products
// @Security BearerAuth
// @Router /api/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.GetInt64("userID")
	role := c.GetString("role")

	// 1. Ownership Check: Fetch the product first
	prod, err := h.client.GetProduct(c.Request.Context(), &pb.GetProductRequest{Id: id})
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Product not found")
		return
	}

	// Only Admin or the specific Seller who owns it can update it
	if role != "admin" && prod.Product.SellerId != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to update this product")
		return
	}

	var reqBody dto.ProductReq
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	_, err = h.client.UpdateProduct(c.Request.Context(), &pb.UpdateProductRequest{
		Id:          id,
		Name:        reqBody.Name,
		Description: reqBody.Description,
		Category:    reqBody.Category,
		Price:       float64(reqBody.Price),
		ImageUrl:    reqBody.ImageURL,
		IsActive:    true,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update product")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Product updated successfully", gin.H{"id": id})
}

// @Summary Delete a product
// @Tags Products
// @Security BearerAuth
// @Router /api/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.GetInt64("userID")
	role := c.GetString("role")

	// 1. Ownership Check
	prod, err := h.client.GetProduct(c.Request.Context(), &pb.GetProductRequest{Id: id})
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Product not found")
		return
	}

	if role != "admin" && prod.Product.SellerId != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to delete this product")
		return
	}

	_, err = h.client.DeleteProduct(c.Request.Context(), &pb.DeleteProductRequest{Id: id})
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Product deleted successfully", gin.H{"id": id})
}