package usecase

import (
	"context"
	"ecommerce/cart-service/internal/domain"
	"ecommerce/pb"
)

type cartUseCase struct {
	repo              domain.CartRepository
	productGrpcClient pb.ProductServiceClient
}

// NewRedisCartRepository creates a new Redis repo
func NewCartUseCase(repo domain.CartRepository, productGrpcClient pb.ProductServiceClient) domain.CartUseCase {
	return &cartUseCase{
		repo:              repo,
		productGrpcClient: productGrpcClient,
	}
}

func (u *cartUseCase) GetCart(ctx context.Context, userID int64) ([]domain.CartItemResponse, error) {
	items, err := u.repo.GetCart(ctx, userID)
	if err != nil || len(items) == 0 {
		return nil, err
	}

	var productIDs []int64
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}

	productRes, err := u.productGrpcClient.GetProductsBatch(ctx, &pb.GetProductsBatchRequest{
		ProductIds: productIDs,
	})
	if err != nil {
		return nil, err
	}

	var fullCart []domain.CartItemResponse
	for _, item := range items {
		if prod, exists := productRes.Products[item.ProductID]; exists {
			fullCart = append(fullCart, domain.CartItemResponse{
				ProductID:      item.ProductID,
				Quantity:       item.Quantity,
				Name:           prod.Name,
				Price:          prod.Price,
				Subtotal:       prod.Price * float64(item.Quantity),
				ImageURL:       prod.ImageUrl,
				SellerShopName: prod.SellerShopName,
				SellerID:       prod.SellerId,
			})
		} else {
			_ = u.repo.RemoveItem(context.Background(), userID, item.ProductID)
		}
	}

	return fullCart, nil
}

func (u *cartUseCase) AddToCart(ctx context.Context, userID int64, productID int64, quantity int32) error {
	return u.repo.AddToCart(ctx, userID, productID, quantity)
}

func (u *cartUseCase) RemoveItem(ctx context.Context, userID int64, productID int64) error {
	return u.repo.RemoveItem(ctx, userID, productID)
}

func (u *cartUseCase) ClearCart(ctx context.Context, userID int64) error {
	return u.repo.ClearCart(ctx, userID)
}

func (u *cartUseCase) GetCartCount(ctx context.Context, userID int64) (int32, error) {
	return u.repo.GetCartCount(ctx, userID)
}
