package domain

import "context"

type Inventory struct {
	ProductID int64
	Stock     int32
}

// Shared Saga Event (From Order Service)
type SagaEvent struct {
	OrderID   int64   `json:"order_id"`
	UserID    int64   `json:"user_id"`
	ProductID int64   `json:"product_id"`
	Amount    float64 `json:"amount"`
}

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

// 2. Matches the Kafka Publisher Envelope!
type ProductEventEnvelope struct {
	EventType string   `json:"event_type"`
	Data      *Product `json:"data"`
}

type InventoryRepository interface {
	InitializeStock(ctx context.Context, productID int64) error
	AdjustStock(ctx context.Context, productID int64, delta int32) error
	GetStock(ctx context.Context, productID int64) (int32, int32, error)
	ReserveStock(ctx context.Context, productID int64, quantity int32) error
	ConfirmStock(ctx context.Context, productID int64, quantity int32) error
	ReleaseStock(ctx context.Context, productID int64, quantity int32) error
}

type KafkaPublisher interface {
	PublishEvent(ctx context.Context, topic string, eventType string, event SagaEvent) error
}

type InventoryUseCase interface {
	AdjustStock(ctx context.Context, productID int64, delta int32) error
	GetStock(ctx context.Context, productID int64) (int32, error)
}