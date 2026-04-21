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
	
	// Create a Tracer specifically for this worker
	tracer := otel.Tracer("inventory-consumer")
	
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			continue
		}

		// Wrap the processing in an anonymous function so defer span.End() runs immediately after each message
		func() {
			eventType := string(m.Key)

			// ==========================================
			// 1. JAEGER TRACING EXTRACTION
			// ==========================================
			// Wrap the incoming Kafka headers in our Carrier
			carrier := tracing.KafkaCarrier{Headers: &m.Headers}

			// Extract the Trace ID to recreate the connected Context
			extractedCtx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

			// Start a new Span linked to the original trace
			spanCtx, span := tracer.Start(extractedCtx, "ConsumeKafkaMessage: "+eventType)
			defer span.End() // Safely closes the span when this anonymous function finishes

			// ==========================================
			// 2. BUSINESS LOGIC (Using spanCtx)
			// ==========================================
			
			// Listen for new products to create empty stock
			if eventType == "product-created" {
				var pEvent domain.ProductEvent
				if err := json.Unmarshal(m.Value, &pEvent); err == nil {
					// Use spanCtx instead of ctx!
					c.repo.InitializeStock(spanCtx, pEvent.ID)
					log.Printf("Initialized stock for new product %d", pEvent.ID)
				}
				return // Exit the anonymous function early
			}

			// Listen for Orders and Payment Failures
			var sEvent domain.SagaEvent
			if err := json.Unmarshal(m.Value, &sEvent); err == nil {

				if eventType == "OrderCreated" {
					// Use spanCtx instead of ctx!
					err := c.repo.ReserveStock(spanCtx, sEvent.ProductID, 1)
					
					if err != nil {
						log.Printf("Out of stock for Order %d! Emitting failure.", sEvent.OrderID)
						// Pass spanCtx to the publisher so the Trace ID continues to the next service!
						c.pub.PublishEvent(spanCtx, "inventory-events", "InventoryFailed", sEvent)
					} else {
						log.Printf("Stock reserved for Order %d! Proceeding to payment.", sEvent.OrderID)
						// Pass spanCtx here too!
						c.pub.PublishEvent(spanCtx, "inventory-events", "InventoryReserved", sEvent)
					}
					
				} else if eventType == "PaymentDeclined" {
					log.Printf("Payment failed for Order %d. Putting stock back.", sEvent.OrderID)
					// Use spanCtx!
					c.repo.Restock(spanCtx, sEvent.ProductID, 1)
				}
			}
		}()
	}
}