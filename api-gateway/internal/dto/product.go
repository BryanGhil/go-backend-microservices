package dto

// ProductReq is the actual JSON body we expect from the frontend
type ProductReq struct {
	Name        string  `json:"name" binding:"required" example:"Wireless Mouse"`
	Price       float64 `json:"price" binding:"required" example:"49.99"`
	Description string  `json:"description" binding:"required" example:"Best mouse in the World"`
	Category    string  `json:"category" example:"Gaming"`
	ImageURL    string  `json:"image_url" example:"www.google.com"`
}

// ProductData represents a single product shape
type ProductData struct {
	ID          int64   `json:"id" example:"1"`
	Name        string  `json:"name" example:"Wireless Mouse"`
	Price       float32 `json:"price" example:"49.99"`
	Description string  `json:"description" example:"Best mouse in the world"`
}

// ProductResponse represents a success returning one product
type ProductResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Product found"`
	Data    ProductData `json:"data"`
}

// ProductListResponse represents a success returning an array of products
type ProductListResponse struct {
	Success bool          `json:"success" example:"true"`
	Message string        `json:"message" example:"Products retrieved successfully"`
	Data    []ProductData `json:"data"`
}
