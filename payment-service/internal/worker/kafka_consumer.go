package worker

import (
	"context"
	"ecommerce/payment-service/internal/domain"
	"ecommerce/payment-service/pkg/tracing"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type PaymentConsumer struct {
	reader *kafka.Reader
	uc     domain.PaymentUseCase
	pub    domain.KafkaPublisher
}

func NewPaymentConsumer(r *kafka.Reader, uc domain.PaymentUseCase, pub domain.KafkaPublisher) *PaymentConsumer {
	return &PaymentConsumer{reader: r, uc: uc, pub: pub}
}

func (c *PaymentConsumer) Start(ctx context.Context) {
	log.Println("Payment Service: Listening for Inventory events...")
	tracer := otel.Tracer("payment-consumer")
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			continue
		}

		func() {
			// 1. Extract the Jaeger Trace ID from the incoming Kafka message
			linkedCtx := tracing.ExtractKafkaContext(context.Background(), m.Headers)

			// 2. Start the span using the linked context
			spanCtx, span := tracer.Start(linkedCtx, "ConsumeKafkaMessage: "+string(m.Key))
			defer span.End() // Safely closes when the function finishes

			// 3. Business Logic
			if string(m.Key) == "InventoryReserved" {
				var event domain.SagaEvent
				json.Unmarshal(m.Value, &event)

				log.Printf("Processing payment for Order %d ($%.2f)...", event.OrderID, event.Amount)
				
				// CRITICAL: Pass spanCtx down to the UseCase so the DB query gets traced!
				success, _ := c.uc.ProcessPayment(spanCtx, event)

				if success {
					log.Printf("Payment SUCCESS for Order %d", event.OrderID)
					// CRITICAL: Pass spanCtx to the publisher so it injects the Trace ID into the next message!
					c.pub.PublishEvent(spanCtx, "payment-events", "PaymentProcessed", event)
				} else {
					log.Printf("Payment DECLINED for Order %d. Initiating Rollback.", event.OrderID)
					// CRITICAL: Pass spanCtx here too!
					c.pub.PublishEvent(spanCtx, "payment-events", "PaymentDeclined", event)
				}
			}
		}() // Execute the anonymous function
	}
}
