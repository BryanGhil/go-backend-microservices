package repository

import (
	"context"
	"database/sql"
	"errors"
	"ecommerce/inventory-service/internal/domain"
)

type pgInventoryRepo struct{ DB *sql.DB }

func NewPostgresInventoryRepo(db *sql.DB) domain.InventoryRepository { return &pgInventoryRepo{DB: db} }

// Called when a new product is created
func (r *pgInventoryRepo) InitializeStock(ctx context.Context, productID int64) error {
	// FIX: Table is 'inventories', column is 'stock_quantity'
	_, err := r.DB.ExecContext(ctx, "INSERT INTO inventories (product_id, stock_quantity) VALUES ($1, 0)", productID)
	return err
}

// Called by Seller manually
func (r *pgInventoryRepo) AddStock(ctx context.Context, productID int64, quantity int32) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE inventories SET stock_quantity = stock_quantity + $1 WHERE product_id = $2", quantity, productID)
	return err
}

func (r *pgInventoryRepo) GetStock(ctx context.Context, productID int64) (int32, error) {
	var stock int32
	err := r.DB.QueryRowContext(ctx, "SELECT stock_quantity FROM inventories WHERE product_id = $1", productID).Scan(&stock)
	return stock, err
}

// THE SAGA DEDUCTION - Prevents double selling
func (r *pgInventoryRepo) ReserveStock(ctx context.Context, productID int64, quantity int32) error {
	res, err := r.DB.ExecContext(ctx, "UPDATE inventories SET stock_quantity = stock_quantity - $1 WHERE product_id = $2 AND stock_quantity >= $1", quantity, productID)
	if err != nil { return err }
	
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("out of stock or product not found")
	}
	return nil
}

// COMPENSATING TRANSACTION - Puts it back if payment fails
func (r *pgInventoryRepo) Restock(ctx context.Context, productID int64, quantity int32) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE inventories SET stock_quantity = stock_quantity + $1 WHERE product_id = $2", quantity, productID)
	return err
}