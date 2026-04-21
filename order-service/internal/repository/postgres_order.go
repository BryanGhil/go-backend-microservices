package repository

import (
	"context"
	"database/sql"
	"ecommerce/order-service/internal/domain"
)

type pgOrderRepo struct{ DB *sql.DB }

func NewPostgresOrderRepo(db *sql.DB) domain.OrderRepository { return &pgOrderRepo{DB: db} }

func (r *pgOrderRepo) Create(ctx context.Context, o *domain.Order) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, "INSERT INTO orders (user_id, product_id, amount, status) VALUES ($1, $2, $3, $4) RETURNING id", 
		o.UserID, o.ProductID, o.Amount, o.Status).Scan(&id)
	return id, err
}

func (r *pgOrderRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE id = $2", status, id)
	return err
}

func (r *pgOrderRepo) GetStatus(ctx context.Context, id int64) (string, error) {
	var status string
	err := r.DB.QueryRowContext(ctx, "SELECT status FROM orders WHERE id = $1", id).Scan(&status)
	return status, err
}