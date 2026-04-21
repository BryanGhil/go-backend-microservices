package domain
import "context"

type Order struct {
	ID        int64
	UserID    int64
	ProductID int64
	Amount    float64
	Status    string // PENDING, COMPLETED, CANCELLED
}

// Shared event struct for Kafka
type SagaEvent struct {
	OrderID   int64   `json:"order_id"`
	UserID    int64   `json:"user_id"`
	ProductID int64   `json:"product_id"`
	Amount    float64 `json:"amount"`
}

type OrderRepository interface {
	Create(ctx context.Context, o *Order) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	GetStatus(ctx context.Context, id int64) (string, error)
}

type KafkaPublisher interface {
	PublishEvent(ctx context.Context, topic string, eventType string, event SagaEvent) error
}

type OrderUseCase interface {
	Checkout(ctx context.Context, userID, productID int64, amount float64) (int64, error)
	GetStatus(ctx context.Context, orderID int64) (string, error)
	HandleSagaCallback(ctx context.Context, eventType string, event SagaEvent) error
}