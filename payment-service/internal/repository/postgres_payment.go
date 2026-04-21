package repository

import (
	"context"
	"database/sql"
	"ecommerce/payment-service/internal/domain"
)

type pgPaymentRepo struct{ DB *sql.DB }

func NewPostgresPaymentRepo(db *sql.DB) domain.PaymentRepository { return &pgPaymentRepo{DB: db} }

func (r *pgPaymentRepo) SaveTransaction(ctx context.Context, p *domain.Payment) error {
	_, err := r.DB.ExecContext(ctx, "INSERT INTO payments (order_id, amount, status) VALUES ($1, $2, $3)", 
		p.OrderID, p.Amount, p.Status)
	return err
}

func (r *pgPaymentRepo) GetStatusByOrderID(ctx context.Context, orderID int64) (string, error) {
	var status string
	err := r.DB.QueryRowContext(ctx, "SELECT status FROM payments WHERE order_id = $1 ORDER BY id DESC LIMIT 1", orderID).Scan(&status)
	return status, err
}