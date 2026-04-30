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
			// Extract trace context
			carrier := tracing.KafkaCarrier{Headers: &m.Headers}
			extractedCtx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

			// Start span
			spanCtx, span := tracer.Start(extractedCtx, "ConsumeKafkaMessage")
			defer span.End()

			// ==========================================
			// 1. PRODUCT EVENTS (JSON Envelope Format)
			// ==========================================
			var envelope domain.ProductEventEnvelope
			// Try to unmarshal into the Envelope first
			if err := json.Unmarshal(m.Value, &envelope); err == nil && envelope.EventType != "" {
				if envelope.EventType == "product_created" {
					c.repo.InitializeStock(spanCtx, envelope.Data.ID)
					log.Printf("Initialized stock for new product %d", envelope.Data.ID)
				}
				return // Exit early
			}

			// ==========================================
			// 2. ORDER SAGA EVENTS (Old direct JSON format)
			// ==========================================
			eventType := string(m.Key)
			var sEvent domain.SagaEvent
			if err := json.Unmarshal(m.Value, &sEvent); err == nil {

				switch eventType {
				case "OrderCreated":
					err := c.repo.ReserveStock(spanCtx, sEvent.ProductID, 1)

					if err != nil {
						log.Printf("Out of stock for Order %d! Emitting failure.", sEvent.OrderID)
						c.pub.PublishEvent(spanCtx, "inventory-events", "InventoryFailed", sEvent)
					} else {
						log.Printf("Stock reserved for Order %d! Proceeding to payment.", sEvent.OrderID)
						c.pub.PublishEvent(spanCtx, "inventory-events", "InventoryReserved", sEvent)
					}

				case "PaymentDeclined":
					log.Printf("Payment failed for Order %d. Putting stock back.", sEvent.OrderID)
					c.repo.Restock(spanCtx, sEvent.ProductID, 1)

				default:
					// Optional: Catch any other events you don't care about yet
					log.Printf("Received unhandled saga event: %s", eventType)
				}
			}
		}()
	}
}
