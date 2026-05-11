package repository

import (
	"context"
	"database/sql"
	"ecommerce/product-service/internal/domain"
	"fmt"
	"strings"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type postgresProductRepo struct {
	DB *sql.DB
}

func NewPostgresProductRepo(db *sql.DB) domain.ProductRepository {
	return &postgresProductRepo{DB: db}
}

func (r *postgresProductRepo) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	// COALESCE prevents Go's sql package from panicking if a column is NULL
	query := `
		SELECT id, seller_id, name, COALESCE(description, ''), COALESCE(category, ''), price, COALESCE(image_url, ''), is_active 
		FROM products 
		WHERE id = $1`

	var p domain.Product
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.SellerID, &p.Name, &p.Description, &p.Category, &p.Price, &p.ImageURL, &p.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *postgresProductRepo) Create(ctx context.Context, p *domain.Product) (int64, error) {
	query := `
		INSERT INTO products (seller_id, name, description, category, price, image_url, is_active, seller_shop_name) 
		VALUES ($1, $2, $3, $4, $5, $6, true, $7) 
		RETURNING id`

	var id int64
	err := r.DB.QueryRowContext(ctx, query, 
		p.SellerID, p.Name, p.Description, p.Category, p.Price, p.ImageURL, p.SellerShopName,
	).Scan(&id)
	
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *postgresProductRepo) Update(ctx context.Context, p *domain.Product) error {
	query := `
		UPDATE products 
		SET name = $1, description = $2, category = $3, price = $4, image_url = $5, is_active = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7`

	result, err := r.DB.ExecContext(ctx, query, 
		p.Name, p.Description, p.Category, p.Price, p.ImageURL, p.IsActive, p.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *postgresProductRepo) Delete(ctx context.Context, id int64) error {
	// Implement "Soft Delete" instead of permanently deleting
	query := `UPDATE products SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

func (r *postgresProductRepo) ListProducts(ctx context.Context, limit, offset int32, sellerID int64, category string) ([]*domain.Product, int64, error) {
	// 1. Build the dynamic WHERE clause
	whereClauses := []string{"is_active = true"} // Only show active products
	args := []interface{}{}
	argID := 1

	if sellerID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("seller_id = $%d", argID))
		args = append(args, sellerID)
		argID++
	}

	if category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argID))
		args = append(args, category)
		argID++
	}

	whereString := strings.Join(whereClauses, " AND ")

	// 2. Execute the COUNT query first for pagination
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM products WHERE %s`, whereString)
	var totalCount int64
	err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// 3. Execute the SELECT query with LIMIT and OFFSET
	selectQuery := fmt.Sprintf(`
		SELECT id, seller_id, name, COALESCE(description, ''), COALESCE(category, ''), price, COALESCE(image_url, '')
		FROM products 
		WHERE %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d`, whereString, argID, argID+1)
	
	args = append(args, limit, offset)

	rows, err := r.DB.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.Category, &p.Price, &p.ImageURL); err != nil {
			return nil, 0, err
		}
		p.IsActive = true // Forced by the WHERE clause
		products = append(products, &p)
	}

	return products, totalCount, nil
}

func (r *postgresProductRepo) UpdateSellerShopName(ctx context.Context, sellerID int64, shopName string) error {
    query := `
        UPDATE products 
        SET seller_shop_name = $1, updated_at = CURRENT_TIMESTAMP
        WHERE seller_id = $2`

    result, err := r.DB.ExecContext(ctx, query, shopName, sellerID)
    if err != nil {
        return err
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return sql.ErrNoRows 
    }

    return nil
}

func (r *postgresProductRepo) GetProductsBatch(ctx context.Context, productIDs []int64) (map[int64]*domain.Product, error) {
	query := `
		SELECT id, seller_id, name, COALESCE(description, ''), COALESCE(category, ''), price, COALESCE(image_url, ''), seller_shop_name, seller_id
		FROM products 
		WHERE id = ANY($1) AND is_active = true
	`
	
	rows, err := r.DB.QueryContext(ctx, query, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 2. Build the Map
	productMap := make(map[int64]*domain.Product)
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.SellerID, &p.Name, &p.Description, &p.Category, &p.Price, &p.ImageURL, &p.SellerShopName, &p.SellerID); err != nil {
			return nil, err
		}
		productMap[p.ID] = &p
	}
	
	return productMap, nil
}