package usecase

import (
	"context"
	"ecommerce/order-service/internal/domain"
	"ecommerce/pb"
	"errors"

	"github.com/google/uuid"
)

type orderUC struct {
	repo          domain.OrderRepository
	pub           domain.KafkaPublisher
	productClient pb.ProductServiceClient
}

func NewOrderUseCase(r domain.OrderRepository, p domain.KafkaPublisher, pc pb.ProductServiceClient) domain.OrderUseCase {
	return &orderUC{repo: r, pub: p, productClient: pc}
}

func (u *orderUC) Checkout(ctx context.Context, userID int64, items []domain.CheckoutItem) (string, error) {
	var productIDs []int64
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}

	realProductsResp, err := u.productClient.GetProductsBatch(ctx, &pb.GetProductsBatchRequest{
		ProductIds: productIDs,
	})
	if err != nil {
		return "", errors.New("failed to validate products with product service")
	}

	realProducts := realProductsResp.GetProducts()

	correlationID := uuid.New().String()
	groupedItems := make(map[int64][]domain.CheckoutItem)
	var sagaItems []domain.SagaItem
	var grandTotal float64

	// 3. Security Override Loop
	for _, item := range items {
		realProduct, exists := realProducts[item.ProductID]
		if !exists {
			return "", errors.New("invalid product ID in cart")
		}

		item.Price = realProduct.Price
		item.SellerID = realProduct.SellerId 

		groupedItems[item.SellerID] = append(groupedItems[item.SellerID], item)
		sagaItems = append(sagaItems, domain.SagaItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
		grandTotal += (item.Price * float64(item.Quantity))
	}

	// 3. Save everything to the database atomically
	err = u.repo.CreateOrderGroupTx(ctx, userID, correlationID, groupedItems)
	if err != nil {
		return "", err
	}

	// 4. Publish ONE Saga Event for the entire checkout
	// Inventory Service will loop through the Items array to deduct stock.
	// Payment Service will charge the TotalAmount based on the CorrelationID.
	event := domain.SagaEvent{
		CorrelationID: correlationID,
		UserID:        userID,
		TotalAmount:   grandTotal,
		Items:         sagaItems,
	}
	u.pub.PublishEvent(ctx, "order-events", "OrderCreated", event)

	return correlationID, nil
}

func (u *orderUC) GetStatus(ctx context.Context, orderID int64) (string, error) {
	return u.repo.GetStatus(ctx, orderID)
}

func (u *orderUC) HandleSagaCallback(ctx context.Context, eventType string, event domain.SagaEvent) error {
	// Notice we now update by CorrelationID, so ALL orders in this checkout are updated simultaneously!
	if eventType == "PaymentProcessed" {
		return u.repo.UpdateStatusByCorrelationID(ctx, event.CorrelationID, "COMPLETED")
	} else if eventType == "PaymentDeclined" || eventType == "InventoryFailed" {
		return u.repo.UpdateStatusByCorrelationID(ctx, event.CorrelationID, "CANCELLED")
	}
	return nil
}

func (u *orderUC) GetUserOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return u.repo.GetUserOrders(ctx, userID)
}