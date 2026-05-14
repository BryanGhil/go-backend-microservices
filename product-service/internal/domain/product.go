package domain

import "context"

type Product struct {
	ID             int64   `json:"id"`
	SellerID       int64   `json:"seller_id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Category       string  `json:"category"`
	Price          float64 `json:"price"`
	ImageURL       string  `json:"image_url"`
	IsActive       bool    `json:"is_active"`
	SellerShopName string  `json:"seller_shop_name"`
	Stock          int32   `json:"stock"`
}

// The interfaces dictate how the layers talk to each other
type ProductRepository interface {
	GetByID(ctx context.Context, id int64) (*Product, error)
	Create(ctx context.Context, p *Product) (int64, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id int64) error
	ListProducts(ctx context.Context, limit, offset int32, sellerID int64, category string) ([]*Product, int64, error)
	UpdateSellerShopName(ctx context.Context, sellerID int64, shopName string) error
	GetProductsBatch(ctx context.Context, productIDs []int64) (map[int64]*Product, error)
}

type ProductUseCase interface {
	GetProduct(ctx context.Context, id int64) (*Product, error)
	CreateProduct(ctx context.Context, p *Product) (int64, error)
	UpdateProduct(ctx context.Context, p *Product) error
	DeleteProduct(ctx context.Context, id int64) error
	ListProducts(ctx context.Context, limit, offset int32, sellerID int64, category string) ([]*Product, int64, error)
	GetSellerDashboardProducts(ctx context.Context, limit, offset int32, sellerID int64) ([]*Product, int64, error)
	GetProductsBatch(ctx context.Context, productIDs []int64) (map[int64]*Product, error)
}

// Add this below your existing interfaces
type ProductEventPublisher interface {
	PublishProductCreated(ctx context.Context, p *Product) error
}
