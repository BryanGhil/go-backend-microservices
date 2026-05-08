package repository

import (
	"context"
	"database/sql"
	"ecommerce/inventory-service/internal/domain"
	"errors"

	"github.com/lib/pq"
)

type pgInventoryRepo struct{ DB *sql.DB }

func NewPostgresInventoryRepo(db *sql.DB) domain.InventoryRepository { return &pgInventoryRepo{DB: db} }

// 1. Initialize (Called when Product is created)
func (r *pgInventoryRepo) InitializeStock(ctx context.Context, productID int64) error {
	_, err := r.DB.ExecContext(ctx, "INSERT INTO inventories (product_id, stock_quantity, reserved_quantity) VALUES ($1, 0, 0)", productID)
	return err
}

// 2. Adjust Stock (Called by Seller manually - Handles positive AND negative numbers)
func (r *pgInventoryRepo) AdjustStock(ctx context.Context, productID int64, delta int32) error {
	// The MAGIC RULE: (stock_quantity + delta) MUST be >= reserved_quantity
	query := `
		UPDATE inventories 
		SET stock_quantity = stock_quantity + $1 
		WHERE product_id = $2 
		AND (stock_quantity + $1) >= reserved_quantity
	`
	res, err := r.DB.ExecContext(ctx, query, delta, productID)
	if err != nil {
		return err
	}
	
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("cannot decrease stock below currently reserved amount")
	}
	return nil
}

// ==========================================
// THE SAGA PATTERN (3-Step Order Lifecycle)
// ==========================================

// STEP 1: Reserve (Order Created - Waiting for Payment)
func (r *pgInventoryRepo) ReserveStock(ctx context.Context, productID int64, quantity int32) error {
	// Check if Available (stock - reserved) >= requested quantity
	query := `
		UPDATE inventories 
		SET reserved_quantity = reserved_quantity + $1 
		WHERE product_id = $2 
		AND (stock_quantity - reserved_quantity) >= $1
	`
	res, err := r.DB.ExecContext(ctx, query, quantity, productID)
	if err != nil { 
		return err 
	}
	
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("out of stock or product not found")
	}
	return nil
}

// STEP 2a: Confirm (Payment Success - Item is officially sold)
func (r *pgInventoryRepo) ConfirmStock(ctx context.Context, productID int64, quantity int32) error {
	// Deduct from BOTH columns, because the item has left the building
	query := `
		UPDATE inventories 
		SET stock_quantity = stock_quantity - $1,
		    reserved_quantity = reserved_quantity - $1 
		WHERE product_id = $2 
		AND reserved_quantity >= $1
	`
	_, err := r.DB.ExecContext(ctx, query, quantity, productID)
	return err
}

// STEP 2b: Release (Payment Failed or Timeout - Unlock the items)
func (r *pgInventoryRepo) ReleaseStock(ctx context.Context, productID int64, quantity int32) error {
	// Only deduct from reserved, putting it back into the "Available" pool
	query := `
		UPDATE inventories 
		SET reserved_quantity = reserved_quantity - $1 
		WHERE product_id = $2 
		AND reserved_quantity >= $1
	`
	_, err := r.DB.ExecContext(ctx, query, quantity, productID)
	return err
}

// 4. Read Data
func (r *pgInventoryRepo) GetStock(ctx context.Context, productID int64) (int32, int32, error) {
	var totalStock, reserved int32
	query := "SELECT stock_quantity, reserved_quantity FROM inventories WHERE product_id = $1"
	err := r.DB.QueryRowContext(ctx, query, productID).Scan(&totalStock, &reserved)
	
	// Returns both total AND reserved, so your UseCase can calculate "Available"
	return totalStock, reserved, err 
}

func (r *pgInventoryRepo) GetStocksBatch(ctx context.Context, productIDs []int64) (map[int64]int32, error) {
	query := `
		SELECT product_id, stock_quantity, reserved_quantity 
		FROM inventories 
		WHERE product_id = ANY($1)
	`
	
	rows, err := r.DB.QueryContext(ctx, query, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 2. Build the Map
	stockMap := make(map[int64]int32)
	for rows.Next() {
		var pID int64
		var total, reserved int32
		if err := rows.Scan(&pID, &total, &reserved); err == nil {
			stockMap[pID] = total - reserved // The exact Available stock!
		}
	}
	
	return stockMap, nil
}