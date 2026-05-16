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
			defer span.End()

			// 3. Business Logic
			if string(m.Key) == "InventoryReserved" {
				var event domain.SagaEvent
				json.Unmarshal(m.Value, &event)

				// FIX: Update logs to use CorrelationID and TotalAmount
				log.Printf("Processing payment for Checkout %s ($%.2f)...", event.CorrelationID, event.TotalAmount)
				
				// Make sure your ProcessPayment UseCase is updated to handle the new event struct!
				success, _ := c.uc.ProcessPayment(spanCtx, event)

				if success {
					log.Printf("Payment SUCCESS for Checkout %s", event.CorrelationID)
					c.pub.PublishEvent(spanCtx, "payment-events", "PaymentProcessed", event)
				} else {
					log.Printf("Payment DECLINED for Checkout %s. Initiating Rollback.", event.CorrelationID)
					c.pub.PublishEvent(spanCtx, "payment-events", "PaymentDeclined", event)
				}
			}
		}() 
	}
}