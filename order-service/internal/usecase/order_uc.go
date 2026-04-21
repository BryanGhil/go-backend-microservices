package usecase

import (
	"context"
	"ecommerce/order-service/internal/domain"
)

type orderUC struct {
	repo domain.OrderRepository
	pub  domain.KafkaPublisher
}

func NewOrderUseCase(r domain.OrderRepository, p domain.KafkaPublisher) domain.OrderUseCase {
	return &orderUC{repo: r, pub: p}
}

func (u *orderUC) Checkout(ctx context.Context, userID, productID int64, amount float64) (int64, error) {
	// 1. Save as PENDING
	order := &domain.Order{UserID: userID, ProductID: productID, Amount: amount, Status: "PENDING"}
	id, err := u.repo.Create(ctx, order)
	if err != nil { return 0, err }

	// 2. Publish to Kafka to trigger Inventory Service
	event := domain.SagaEvent{OrderID: id, UserID: userID, ProductID: productID, Amount: amount}
	u.pub.PublishEvent(ctx, "order-events", "OrderCreated", event)

	return id, nil
}

func (u *orderUC) GetStatus(ctx context.Context, orderID int64) (string, error) {
	return u.repo.GetStatus(ctx, orderID)
}

// 3. This is called by the Kafka Consumer when other services reply
func (u *orderUC) HandleSagaCallback(ctx context.Context, eventType string, event domain.SagaEvent) error {
	if eventType == "PaymentProcessed" {
		return u.repo.UpdateStatus(ctx, event.OrderID, "COMPLETED")
	} else if eventType == "PaymentDeclined" || eventType == "InventoryFailed" {
		return u.repo.UpdateStatus(ctx, event.OrderID, "CANCELLED")
	}
	return nil
}