package repository

import (
	"context"
	"database/sql"
	"ecommerce/user-service/internal/domain"
)

type userRepo struct { DB *sql.DB }

func NewPostgresUserRepo(db *sql.DB) domain.UserRepository {
	return &userRepo{DB: db}
}

func (r *userRepo) CreateUser(ctx context.Context, email, passwordHash string) (int64, error) {
	var id int64
	err := r.DB.QueryRowContext(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, passwordHash).Scan(&id)
	return id, err
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.DB.QueryRowContext(ctx, "SELECT id, email, password_hash FROM users WHERE email = $1", email).Scan(&u.ID, &u.Email, &u.PasswordHash)
	return &u, err
}

func (r *userRepo) UpdateProfile(ctx context.Context, id int64, name, phone, address string) error {
	_, err := r.DB.ExecContext(ctx, "UPDATE users SET full_name = $1, phone = $2, address = $3 WHERE id = $4", name, phone, address, id)
	return err
}