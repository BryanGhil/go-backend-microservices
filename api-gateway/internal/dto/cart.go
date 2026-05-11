package dto

type AddToCartReq struct {
	ProductID int64 `json:"product_id" binding:"required" example:"105"`
	Quantity  int32 `json:"quantity" binding:"required" example:"1"`
}