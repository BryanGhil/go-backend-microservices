package domain
import "context"

type Payment struct {
	ID            int64
	CorrelationID string  // FIX: Replaced OrderID with CorrelationID
	Amount        float64
	Status        string  // SUCCESS, DECLINED
}

type SagaItem struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type SagaEvent struct {
	CorrelationID string     `json:"correlation_id"` 
	UserID        int64      `json:"user_id"`
	TotalAmount   float64    `json:"total_amount"`
	Items         []SagaItem `json:"items"`
}

type PaymentRepository interface {
	SaveTransaction(ctx context.Context, p *Payment) error
    // FIX: Update to expect a string
	GetStatusByCorrelationID(ctx context.Context, correlationID string) (string, error) 
}

type KafkaPublisher interface {
	PublishEvent(ctx context.Context, topic string, eventType string, event SagaEvent) error
}

type PaymentUseCase interface {
	ProcessPayment(ctx context.Context, event SagaEvent) (bool, error)
    // FIX: Update to expect a string
	GetStatus(ctx context.Context, correlationID string) (string, error) 
}