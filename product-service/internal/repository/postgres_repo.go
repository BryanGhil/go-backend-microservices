package repository

import (
	"context"
	"database/sql"
	"ecommerce/product-service/internal/domain"
	_ "github.com/lib/pq" // Postgres driver
)

type postgresProductRepo struct {
	DB *sql.DB
}

func NewPostgresProductRepo(db *sql.DB) domain.ProductRepository {
	return &postgresProductRepo{DB: db}
}

func (r *postgresProductRepo) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	// PostgreSQL uses $1, $2, etc. for placeholders
	query := `SELECT id, name, price FROM products WHERE id = $1`
	
	var p domain.Product
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Name, &p.Price)
	if err != nil {
		return nil, err
	}
	
	return &p, nil
}

func (r *postgresProductRepo) Create(ctx context.Context, p *domain.Product) (int64, error) {
	query := `INSERT INTO products (name, price) VALUES ($1, $2) RETURNING id`
	
	var id int64
	err := r.DB.QueryRowContext(ctx, query, p.Name, p.Price).Scan(&id)
	if err != nil {
		return 0, err
	}
	
	return id, nil
}

func (r *postgresProductRepo) Update(ctx context.Context, p *domain.Product) error {
	query := `UPDATE products SET name = $1, price = $2 WHERE id = $3`
	
	// ExecContext doesn't return rows, just the result (rows affected)
	result, err := r.DB.ExecContext(ctx, query, p.Name, p.Price, p.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	// We return a specific error so the gRPC handler knows it was a 404
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *postgresProductRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

func (r *postgresProductRepo) GetAll(ctx context.Context) ([]*domain.Product, error) {
	query := `SELECT id, name, price FROM products`
	
	// QueryContext returns multiple rows
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*domain.Product

	// Loop through the results
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	return products, nil
}