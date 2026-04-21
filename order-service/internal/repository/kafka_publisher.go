package repository

import (
	"context"
	"ecommerce/order-service/internal/domain"
	"ecommerce/order-service/pkg/tracing"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type kafkaPub struct{ writer *kafka.Writer }

func NewKafkaPublisher(w *kafka.Writer) domain.KafkaPublisher { return &kafkaPub{writer: w} }

func (k *kafkaPub) PublishEvent(ctx context.Context, topic string, eventType string, event domain.SagaEvent) error {
	bytes, _ := json.Marshal(event)

	// 1. Create empty Kafka Headers
	var headers []kafka.Header

	// 2. Wrap them in our Carrier
	carrier := tracing.KafkaCarrier{Headers: &headers}

	// 3. INJECT the Trace ID from the Go context into the Kafka Headers!
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	
	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(eventType),
		Value: bytes,
		Headers: headers,
	})
}