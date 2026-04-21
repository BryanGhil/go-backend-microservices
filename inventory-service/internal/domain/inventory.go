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

// Shared Product Event (From Product Service)
type ProductEvent struct {
	ID    int64   `json:"ID"`
	Name  string  `json:"Name"`
	Price float64 `json:"Price"`
}

type InventoryRepository interface {
	InitializeStock(ctx context.Context, productID int64) error
	AddStock(ctx context.Context, productID int64, quantity int32) error
	GetStock(ctx context.Context, productID int64) (int32, error)
	ReserveStock(ctx context.Context, productID int64, quantity int32) error
	Restock(ctx context.Context, productID int64, quantity int32) error // Compensating Transaction
}

type KafkaPublisher interface {
	PublishEvent(ctx context.Context, topic string, eventType string, event SagaEvent) error
}

type InventoryUseCase interface {
	AddStock(ctx context.Context, productID int64, quantity int32) error
	GetStock(ctx context.Context, productID int64) (int32, error)
}