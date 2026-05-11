package domain

import "context"

type CartItem struct {
	ProductID int64
	Quantity  int32
}

type CartItemResponse struct {
	ProductID      int64
	Quantity       int32
	Name           string
	Price          float64
	Subtotal       float64
	ImageURL       string
	SellerID       int64
	SellerShopName string
}

type CartRepository interface {
	AddToCart(ctx context.Context, userID int64, productID int64, quantity int32) error
	RemoveItem(ctx context.Context, userID int64, productID int64) error
	GetCart(ctx context.Context, userID int64) ([]CartItem, error)
	ClearCart(ctx context.Context, userID int64) error
	GetCartCount(ctx context.Context, userID int64) (int32, error)
}

type CartUseCase interface {
	GetCart(ctx context.Context, userID int64) ([]CartItemResponse, error)
	AddToCart(ctx context.Context, userID int64, productID int64, quantity int32) error
	RemoveItem(ctx context.Context, userID int64, productID int64) error
	ClearCart(ctx context.Context, userID int64) error
	GetCartCount(ctx context.Context, userID int64) (int32, error)
}
