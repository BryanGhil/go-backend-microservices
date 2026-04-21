package worker

import (
	"context"
	"ecommerce/order-service/internal/domain"
	"ecommerce/order-service/pkg/tracing"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type SagaConsumer struct {
	reader *kafka.Reader
	uc     domain.OrderUseCase
}

func NewSagaConsumer(r *kafka.Reader, uc domain.OrderUseCase) *SagaConsumer {
	return &SagaConsumer{reader: r, uc: uc}
}

func (c *SagaConsumer) Start(ctx context.Context) {
	log.Println("Order Service: Listening for Saga callbacks...")
	tracer := otel.Tracer("order-consumer")

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

			eventType := string(m.Key)
			var event domain.SagaEvent
			if err := json.Unmarshal(m.Value, &event); err == nil {
				// If we hear from Payment or Inventory, update the order status
				if eventType == "PaymentProcessed" || eventType == "PaymentDeclined" || eventType == "InventoryFailed" {
					c.uc.HandleSagaCallback(spanCtx, eventType, event)
					log.Printf("Saga Step Completed: Order %d is now reacting to %s", event.OrderID, eventType)
				}
			}
		}()

	}
}
