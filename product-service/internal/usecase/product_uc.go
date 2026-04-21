package usecase

import (
	"context"
	"ecommerce/product-service/internal/domain"
	"fmt"
)

type productUseCase struct {
	repo domain.ProductRepository
	publisher domain.ProductEventPublisher
}

func NewProductUseCase(repo domain.ProductRepository, pub domain.ProductEventPublisher) domain.ProductUseCase {
	return &productUseCase{repo: repo, publisher: pub}
}

func (u *productUseCase) GetProduct(ctx context.Context, id int64) (*domain.Product, error) {
	// You can add business logic here (e.g., checking if the product is active)
	return u.repo.GetByID(ctx, id)
}

func (u *productUseCase) CreateProduct(ctx context.Context, name string, price float64) (int64, error) {
	// Simple validation example
	if price < 0 {
		return 0, fmt.Errorf("price cannot be negative")
	}

	product := &domain.Product{
		Name:  name,
		Price: price,
	}

	id, err := u.repo.Create(ctx, product)
	if err != nil {
		return 0, err
	}
	product.ID = id

	// 2. Publish to Kafka (Asynchronous-ish)
	// If this fails, we log it, but we DON'T fail the user's request. 
	// The product is already in Postgres.
	_ = u.publisher.PublishProductCreated(ctx, product) 

	return id, nil
}

func (u *productUseCase) UpdateProduct(ctx context.Context, id int64, name string, price float64) error {
	p := &domain.Product{ID: id, Name: name, Price: price}
	return u.repo.Update(ctx, p)
}

func (u *productUseCase) DeleteProduct(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}

func (u *productUseCase) GetAllProducts(ctx context.Context) ([]*domain.Product, error) {
	return u.repo.GetAll(ctx)
}