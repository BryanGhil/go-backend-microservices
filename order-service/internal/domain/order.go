package domain

import "context"

type Order struct {
	ID            int64
	UserID        int64
	ProductID     int64
	Amount        float64
	Currency      string
	Status        string // PENDING, COMPLETED, CANCELLED
	CorrelationID string
}

type CheckoutItem struct {
	ProductID int64
	SellerID  int64
	Quantity  int
	Price     float64
}

type OrderItem struct {
	ID              int64
	OrderID         int64
	ProductID       int64
	SellerID        int64
	Quantity        int
	PriceAtPurchase float64
	Status          string
}

type Shipment struct {
	ID       int64
	OrderID  int64
	SellerID int64
	Status   string
}

type SagaItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

// Shared event struct for Kafka
type SagaEvent struct {
	CorrelationID string     `json:"correlation_id"` // Replaces OrderID for Saga tracking
	UserID        int64      `json:"user_id"`
	TotalAmount   float64    `json:"total_amount"`
	Items         []SagaItem `json:"items"`
}

type OrderRepository interface {
	CreateOrderGroupTx(ctx context.Context, userID int64, correlationID string, groupedItems map[int64][]CheckoutItem) error
	UpdateStatusByCorrelationID(ctx context.Context, correlationID string, status string) error
	GetStatus(ctx context.Context, orderID int64) (string, error)
	GetUserOrders(ctx context.Context, userID int64) ([]*Order, error)
}

type KafkaPublisher interface {
	PublishEvent(ctx context.Context, topic string, eventType string, event SagaEvent) error
}

type OrderUseCase interface {
	Checkout(ctx context.Context, userID int64, items []CheckoutItem) (string, error) // Returns CorrelationID
	GetStatus(ctx context.Context, orderID int64) (string, error)
	HandleSagaCallback(ctx context.Context, eventType string, event SagaEvent) error
	GetUserOrders(ctx context.Context, userID int64) ([]*Order, error)
}
