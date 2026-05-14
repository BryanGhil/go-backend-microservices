package repository

import (
	"context"
	"database/sql"
	"ecommerce/user-service/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type userRepo struct{ DB *sql.DB }

func NewPostgresUserRepo(db *sql.DB) domain.UserRepository {
	return &userRepo{DB: db}
}

func (r *userRepo) CreateUser(ctx context.Context, u *domain.User) (int64, error) {
	// 1. Start Trace
	tracer := otel.Tracer("user-repository")
	ctx, span := tracer.Start(ctx, "Postgres.CreateUser")
	defer span.End()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to begin transaction")
		return 0, err
	}
	defer tx.Rollback()

	var id int64
	queryUser := `
		INSERT INTO users (email, password_hash, full_name, role, is_active, provider) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err = tx.QueryRowContext(ctx, queryUser,
		u.Email, u.PasswordHash, u.FullName, u.Role, u.IsActive, u.Provider,
	).Scan(&id)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to insert into users table")
		return 0, err
	}

	if u.Role == "seller" && u.SellerProfile != nil {
		querySeller := `INSERT INTO seller_profiles (user_id, shop_name) VALUES ($1, $2)`
		_, err = tx.ExecContext(ctx, querySeller, id, u.SellerProfile.ShopName)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to insert into seller_profiles table")
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to commit transaction")
		return 0, err
	}

	return id, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	tracer := otel.Tracer("user-repository")
	ctx, span := tracer.Start(ctx, "Postgres.GetUserByEmail")
	defer span.End()

	var u domain.User
	query := `SELECT id, email, password_hash, role, is_active FROM users WHERE email = $1`
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch user by email")
	}
	return &u, err
}

// CORRECTED REPOSITORY
func (r *userRepo) GetUserById(ctx context.Context, id int64) (*domain.User, error) {
	tracer := otel.Tracer("user-repository")
	ctx, span := tracer.Start(ctx, "Postgres.GetUserById")
	defer span.End()

	var u domain.User
	var shopName sql.NullString
	var shopDescr sql.NullString

	// FIX: Added shop_name to the SELECT query
	query := `SELECT u.id, u.email, u.password_hash, u.role, u.is_active, u.full_name, u.phone_number, u.address, sp.shop_name, sp.shop_description FROM users u LEFT JOIN seller_profiles sp on u.id = sp.user_id WHERE u.id = $1`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.FullName, &u.Phone, &u.Address, &shopName, &shopDescr,
	)

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if u.Role == "seller" && shopName.Valid {
		u.SellerProfile = &domain.SellerProfile{
			ShopName:        shopName.String,
			ShopDescription: shopDescr.String,
		}
	}

	return &u, nil
}

func (r *userRepo) UpdateProfile(ctx context.Context, u *domain.User) error {
	tracer := otel.Tracer("user-repository")
	ctx, span := tracer.Start(ctx, "Postgres.UpdateProfile")
	defer span.End()

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	queryUser := `UPDATE users SET full_name = $1, phone_number = $2, address = $3 WHERE id = $4`
	_, err = tx.ExecContext(ctx, queryUser, u.FullName, u.Phone, u.Address, u.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update users table")
		return err
	}

	if u.SellerProfile != nil {
		querySeller := `
			UPDATE seller_profiles  
			SET shop_name = $1, shop_description = $2
			WHERE user_id = $3;
			`

		_, err = tx.ExecContext(ctx, querySeller, u.SellerProfile.ShopName, u.SellerProfile.ShopDescription, u.ID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to upsert seller_profiles table")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to commit transaction")
		return err
	}

	return nil
}
