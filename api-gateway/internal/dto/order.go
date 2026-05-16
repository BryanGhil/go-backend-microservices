package dto

type CheckoutItemReq struct {
	ProductID int64   `json:"product_id" binding:"required" example:"105"`
	Quantity  int32   `json:"quantity" binding:"required,min=1" example:"2"`
}

type CheckoutReq struct {
	Items []CheckoutItemReq `json:"items" binding:"required,dive"`
}