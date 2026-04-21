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
	_, err := r.DB.ExecContext(ctx, "INSERT INTO inventory (product_id, stock) VALUES ($1, 0)", productID)
	return err
}

// Called by Seller manually
func (r *pgInventoryRepo) AddStock(ctx context.Context, productID int64, quantity int32) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE inventory SET stock = stock + $1 WHERE product_id = $2", quantity, productID)
	return err
}

func (r *pgInventoryRepo) GetStock(ctx context.Context, productID int64) (int32, error) {
	var stock int32
	err := r.DB.QueryRowContext(ctx, "SELECT stock FROM inventory WHERE product_id = $1", productID).Scan(&stock)
	return stock, err
}

// THE SAGA DEDUCTION - Prevents double selling using "stock >= quantity"
func (r *pgInventoryRepo) ReserveStock(ctx context.Context, productID int64, quantity int32) error {
	res, err := r.DB.ExecContext(ctx, "UPDATE inventory SET stock = stock - $1 WHERE product_id = $2 AND stock >= $1", quantity, productID)
	if err != nil { return err }
	
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("out of stock or product not found")
	}
	return nil
}

// COMPENSATING TRANSACTION - Puts it back if payment fails
func (r *pgInventoryRepo) Restock(ctx context.Context, productID int64, quantity int32) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE inventory SET stock = stock + $1 WHERE product_id = $2", quantity, productID)
	return err
}