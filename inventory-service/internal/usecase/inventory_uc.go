package usecase

import (
	"context"
	"ecommerce/inventory-service/internal/domain"
)

type inventoryUC struct {
	repo domain.InventoryRepository
}

func NewInventoryUseCase(r domain.InventoryRepository) domain.InventoryUseCase {
	return &inventoryUC{repo: r}
}

func (u *inventoryUC) AddStock(ctx context.Context, productID int64, quantity int32) error {
	return u.repo.AddStock(ctx, productID, quantity)
}

func (u *inventoryUC) GetStock(ctx context.Context, productID int64) (int32, error) {
	return u.repo.GetStock(ctx, productID)
}