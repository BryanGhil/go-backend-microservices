package repository

import (
	"context"
	"database/sql"
	"ecommerce/order-service/internal/domain"
)

type pgOrderRepo struct{ DB *sql.DB }

func NewPostgresOrderRepo(db *sql.DB) domain.OrderRepository { return &pgOrderRepo{DB: db} }

func (r *pgOrderRepo) CreateOrderGroupTx(ctx context.Context, userID int64, correlationID string, groupedItems map[int64][]domain.CheckoutItem) error {
	// Start DB Transaction
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Defer a rollback. If the transaction is successfully committed, this does nothing.
	defer tx.Rollback()

	// Loop through each Seller Group
	for sellerID, items := range groupedItems {
		var sellerTotal float64
		for _, item := range items {
			sellerTotal += (item.Price * float64(item.Quantity))
		}

		// 1. Create the Order for this specific Seller
		var orderID int64
		orderQuery := `INSERT INTO orders (user_id, amount, status, correlation_id) VALUES ($1, $2, 'PENDING', $3) RETURNING id`
		err = tx.QueryRowContext(ctx, orderQuery, userID, sellerTotal, correlationID).Scan(&orderID)
		if err != nil {
			return err
		}

		// 2. Create Order Items
		for _, item := range items {
			itemQuery := `INSERT INTO order_items (order_id, product_id, seller_id, quantity, price_at_purchase) VALUES ($1, $2, $3, $4, $5)`
			_, err = tx.ExecContext(ctx, itemQuery, orderID, item.ProductID, item.SellerID, item.Quantity, item.Price)
			if err != nil {
				return err
			}
		}

		// 3. Create initial Shipment tracking record
		shipmentQuery := `INSERT INTO shipments (order_id, seller_id, status) VALUES ($1, $2, 'PENDING')`
		_, err = tx.ExecContext(ctx, shipmentQuery, orderID, sellerID)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
}

func (r *pgOrderRepo) UpdateStatusByCorrelationID(ctx context.Context, correlationID string, status string) error {
	// This updates ALL split orders from the same checkout!
	_, err := r.DB.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE correlation_id = $2", status, correlationID)
	return err
}

func (r *pgOrderRepo) GetStatus(ctx context.Context, id int64) (string, error) {
	var status string
	err := r.DB.QueryRowContext(ctx, "SELECT status FROM orders WHERE id = $1", id).Scan(&status)
	return status, err
}

// Add to domain.OrderRepository interface and implement here:
func (r *pgOrderRepo) GetUserOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	query := `SELECT id, user_id, amount, status, correlation_id FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		err := rows.Scan(&o.ID, &o.UserID, &o.Amount, &o.Status, &o.CorrelationID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}