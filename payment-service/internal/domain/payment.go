package domain
import "context"

type Payment struct {
	ID      int64
	OrderID int64
	Amount  float64
	Status  string // SUCCESS, DECLINED
}

// Shared Saga Event
type SagaEvent struct {
	OrderID   int64   `json:"order_id"`
	UserID    int64   `json:"user_id"`
	ProductID int64   `json:"product_id"`
	Amount    float64 `json:"amount"`
}

type PaymentRepository interface {
	SaveTransaction(ctx context.Context, p *Payment) error
	GetStatusByOrderID(ctx context.Context, orderID int64) (string, error)
}

type KafkaPublisher interface {
	PublishEvent(ctx context.Context, topic string, eventType string, event SagaEvent) error
}

type PaymentUseCase interface {
	ProcessPayment(ctx context.Context, event SagaEvent) (bool, error)
	GetStatus(ctx context.Context, orderID int64) (string, error)
}