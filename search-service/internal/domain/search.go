package domain

import "context"

// 1. Matches the new Product Service fields
type Product struct {
	ID          int64   `json:"id"`
	SellerID    int64   `json:"seller_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
	IsActive    bool    `json:"is_active"`
}

// 2. Matches the Kafka Publisher Envelope we created earlier
type ProductEventEnvelope struct {
	EventType string   `json:"event_type"`
	Data      *Product `json:"data"`
}

type SearchRepository interface {
	IndexProduct(ctx context.Context, p *Product) error
	Search(ctx context.Context, query string, limit, offset int32) ([]*Product, int64, error)
}