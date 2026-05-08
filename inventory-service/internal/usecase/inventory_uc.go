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

// FIX: Renamed to AdjustStock to handle both additions and deductions
func (u *inventoryUC) AdjustStock(ctx context.Context, productID int64, delta int32) error {
	return u.repo.AdjustStock(ctx, productID, delta)
}

// FIX: Calculate "Available" stock for the frontend
func (u *inventoryUC) GetStock(ctx context.Context, productID int64) (int32, error) {
	// The repo now returns two numbers: total physical stock, and reserved stock
	totalStock, reserved, err := u.repo.GetStock(ctx, productID)
	if err != nil {
		return 0, err
	}
	
	// The true number of items we are allowed to sell right now
	availableStock := totalStock - reserved
	
	// If availableStock goes negative due to a data anomaly, return 0
	if availableStock < 0 {
		return 0, nil
	}

	return availableStock, nil
}

func (u *inventoryUC) GetStocksBatch(ctx context.Context, productIDs []int64) (map[int64]int32, error) {
	return u.repo.GetStocksBatch(ctx, productIDs)
}