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
	inventoryGrpcClient pb.InventoryServiceClient
}

func NewProductUseCase(repo domain.ProductRepository, pub domain.ProductEventPublisher, userGrpcClient pb.UserServiceClient, inventoryGrpcClient pb.InventoryServiceClient) domain.ProductUseCase {
	return &productUseCase{repo: repo, publisher: pub, userGrpcClient: userGrpcClient, inventoryGrpcClient: inventoryGrpcClient}
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

func (u *productUseCase) GetSellerDashboardProducts(ctx context.Context, limit, offset int32, sellerID int64) ([]*domain.Product, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	products, count, err := u.repo.ListProducts(ctx, limit, offset, sellerID, "")
	if err != nil {
		return nil, 0, err
	}

	var productIDS []int64
	for _, p := range products {
		productIDS = append(productIDS, p.ID)
	}

	productStock, err := u.inventoryGrpcClient.GetStocksBatch(ctx, &pb.GetStocksBatchRequest{ProductIds: productIDS})
	if err != nil {
		return nil, 0, err
	}

	for _, p := range products {
		if stock, exists := productStock.Stocks[p.ID]; exists {
			p.Stock = stock
		} else {
			p.Stock = 0
		}
	}

	return products, count, err
}

func (u *productUseCase) GetProductsBatch(ctx context.Context, productIDs []int64) (map[int64]*domain.Product, error) {
	return u.repo.GetProductsBatch(ctx, productIDs)
}