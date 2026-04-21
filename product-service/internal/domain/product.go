package domain

import "context"

type Product struct {
	ID    int64
	Name  string
	Price float64
}

// The interfaces dictate how the layers talk to each other
type ProductRepository interface {
	GetByID(ctx context.Context, id int64) (*Product, error)
	Create(ctx context.Context, p *Product) (int64, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id int64) error
	GetAll(ctx context.Context) ([]*Product, error)
}

type ProductUseCase interface {
	GetProduct(ctx context.Context, id int64) (*Product, error)
	CreateProduct(ctx context.Context, name string, price float64) (int64, error)
	UpdateProduct(ctx context.Context, id int64, name string, price float64) error
	DeleteProduct(ctx context.Context, id int64) error
	GetAllProducts(ctx context.Context) ([]*Product, error)
}

// Add this below your existing interfaces
type ProductEventPublisher interface {
	PublishProductCreated(ctx context.Context, p *Product) error
}