package repository

import (
	"context"
	"database/sql"
	"ecommerce/user-service/internal/domain"
)

type userRepo struct{ DB *sql.DB }

func NewPostgresUserRepo(db *sql.DB) domain.UserRepository {
	return &userRepo{DB: db}
}

func (r *userRepo) CreateUser(ctx context.Context, u *domain.User) (int64, error) {
	// Start a transaction
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // Rolls back if tx.Commit() is not called

	var id int64
	queryUser := `
		INSERT INTO users (email, password_hash, full_name, role, is_active, provider) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	
	err = tx.QueryRowContext(ctx, queryUser, 
		u.Email, u.PasswordHash, u.FullName, u.Role, u.IsActive, u.Provider,
	).Scan(&id)
	
	if err != nil {
		return 0, err
	}

	// If it's a seller, insert into the second table
	if u.Role == "seller" && u.SellerProfile != nil {
		querySeller := `INSERT INTO seller_profiles (user_id, shop_name) VALUES ($1, $2)`
		_, err = tx.ExecContext(ctx, querySeller, id, u.SellerProfile.ShopName)
		if err != nil {
			return 0, err
		}
	}

	// Commit the transaction
	return id, tx.Commit()
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	// Added role and is_active to the select query
	query := `SELECT id, email, password_hash, role, is_active FROM users WHERE email = $1`
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive,
	)
	return &u, err
}

func (r *userRepo) UpdateProfile(ctx context.Context, u *domain.User) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryUser := `UPDATE users SET full_name = $1, phone_number = $2, address = $3 WHERE id = $4`
	_, err = tx.ExecContext(ctx, queryUser, u.FullName, u.Phone, u.Address, u.ID)
	if err != nil {
		return err
	}

	// Update seller profile using UPSERT (ON CONFLICT)
	if u.SellerProfile != nil {
		querySeller := `
			INSERT INTO seller_profiles (user_id, shop_name, shop_description) 
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id) DO UPDATE 
			SET shop_name = EXCLUDED.shop_name, 
			    shop_description = EXCLUDED.shop_description, 
			    updated_at = CURRENT_TIMESTAMP`
			    
		_, err = tx.ExecContext(ctx, querySeller, u.ID, u.SellerProfile.ShopName, u.SellerProfile.ShopDescription)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}