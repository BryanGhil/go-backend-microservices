package worker

import (
	"context"
	"encoding/json"
	"log"

	"ecommerce/inventory-service/internal/domain"
	"ecommerce/inventory-service/tracing"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type InventoryConsumer struct {
	reader *kafka.Reader
	repo   domain.InventoryRepository
	pub    domain.KafkaPublisher
}

func NewInventoryConsumer(r *kafka.Reader, repo domain.InventoryRepository, pub domain.KafkaPublisher) *InventoryConsumer {
	return &InventoryConsumer{reader: r, repo: repo, pub: pub}
}

func (c *InventoryConsumer) Start(ctx context.Context) {
	log.Println("Inventory Service: Listening for events...")
	tracer := otel.Tracer("inventory-consumer")

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			continue
		}

		func() {
			carrier := tracing.KafkaCarrier{Headers: &m.Headers}
			extractedCtx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

			spanCtx, span := tracer.Start(extractedCtx, "ConsumeKafkaMessage")
			defer span.End()

			// 1. PRODUCT EVENTS
			var envelope domain.ProductEventEnvelope
			if err := json.Unmarshal(m.Value, &envelope); err == nil && envelope.EventType != "" {
				if envelope.EventType == "product_created" {
					c.repo.InitializeStock(spanCtx, envelope.Data.ID)
					log.Printf("Initialized stock for new product %d", envelope.Data.ID)
				}
				return
			}

			// 2. ORDER SAGA EVENTS
			eventType := string(m.Key)
			var sEvent domain.SagaEvent
			if err := json.Unmarshal(m.Value, &sEvent); err == nil {

				switch eventType {
				case "OrderCreated":
					// FIX: We must loop through ALL items in the cart!
					var reservedItems []domain.SagaItem
					var outOfStock bool

					for _, item := range sEvent.Items {
						err := c.repo.ReserveStock(spanCtx, item.ProductID, int32(item.Quantity))
						if err != nil {
							outOfStock = true
							break // Stop processing if even ONE item fails
						}
						reservedItems = append(reservedItems, item)
					}

					// Partial Rollback Logic
					if outOfStock {
						log.Printf("Out of stock for Checkout %s! Rolling back reserved items.", sEvent.CorrelationID)
						// Release the items we successfully reserved before hitting the failure
						for _, item := range reservedItems {
							c.repo.ReleaseStock(spanCtx, item.ProductID, int32(item.Quantity))
						}
						c.pub.PublishEvent(spanCtx, "inventory-events", "InventoryFailed", sEvent)
					} else {
						log.Printf("All stock reserved for Checkout %s! Proceeding to payment.", sEvent.CorrelationID)
						c.pub.PublishEvent(spanCtx, "inventory-events", "InventoryReserved", sEvent)
					}

				case "PaymentDeclined":
					log.Printf("Payment failed for Checkout %s. Releasing all stock.", sEvent.CorrelationID)
					// FIX: Loop to release all items
					for _, item := range sEvent.Items {
						c.repo.ReleaseStock(spanCtx, item.ProductID, int32(item.Quantity))
					}

				case "PaymentProcessed": // Ensure this matches what Payment service sends!
					log.Printf("Payment success for Checkout %s. Confirming final stock deductions.", sEvent.CorrelationID)
					// FIX: Loop to confirm all items
					for _, item := range sEvent.Items {
						c.repo.ConfirmStock(spanCtx, item.ProductID, int32(item.Quantity))
					}

				default:
					log.Printf("Received unhandled saga event: %s", eventType)
				}
			}
		}()
	}
}