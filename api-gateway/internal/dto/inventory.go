package dto

type AddStockReq struct {
	ProductID int64 `json:"product_id" binding:"required" example:"1"`
	Quantity  int32 `json:"quantity" binding:"required,gt=0" example:"50"` // gt=0 ensures they can't add negative stock
}