package domain
import "context"

type Product struct {
	ID    int64   `json:"ID"` // Must match the JSON sent by the publisher
	Name  string  `json:"Name"`
	Price float64 `json:"Price"`
}

type SearchRepository interface {
	IndexProduct(ctx context.Context, p *Product) error
	Search(ctx context.Context, query string) ([]*Product, error)
}