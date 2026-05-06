package usecase

import (
	"context"
	"ecommerce/pb"
	"ecommerce/product-service/internal/domain"
	"fmt"
)

type productUseCase struct {
	repo      domain.ProductRepository
	publisher domain.ProductEventPublisher
	userGrpcClient pb.UserServiceClient
}

func NewProductUseCase(repo domain.ProductRepository, pub domain.ProductEventPublisher, userGrpcClient pb.UserServiceClient) domain.ProductUseCase {
	return &productUseCase{repo: repo, publisher: pub, userGrpcClient: userGrpcClient}
}

func (u *productUseCase) GetProduct(ctx context.Context, id int64) (*domain.Product, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *productUseCase) CreateProduct(ctx context.Context, p *domain.Product) (int64, error) {
	// Validation
	if p.Price < 0 {
		return 0, fmt.Errorf("price cannot be negative")
	}
	if p.SellerID == 0 {
		return 0, fmt.Errorf("seller ID is required")
	}
	userRes, err := u.userGrpcClient.GetSellerProfile(ctx, &pb.GetSellerProfileReq{
		SellerId: p.SellerID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to validate seller or fetch shop name")
	}

	p.SellerShopName = userRes.ShopName

	id, err := u.repo.Create(ctx, p)
	if err != nil {
		return 0, err
	}
	p.ID = id
	p.IsActive = true // Default to active on creation

	// Publish to Kafka
	_ = u.publisher.PublishProductCreated(ctx, p)

	return id, nil
}

func (u *productUseCase) UpdateProduct(ctx context.Context, p *domain.Product) error {
	if p.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	return u.repo.Update(ctx, p)
}

func (u *productUseCase) DeleteProduct(ctx context.Context, id int64) error {
	// Using the repository's soft delete (setting is_active = false)
	return u.repo.Delete(ctx, id)
}

func (u *productUseCase) ListProducts(ctx context.Context, limit, offset int32, sellerID int64, category string) ([]*domain.Product, int64, error) {
	// Set safe defaults for pagination to prevent massive DB queries
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return u.repo.ListProducts(ctx, limit, offset, sellerID, category)
}