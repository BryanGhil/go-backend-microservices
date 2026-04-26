package repository

import (
	"context"
	"ecommerce/product-service/internal/domain"
	"ecommerce/product-service/pkg/tracing"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type kafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(writer *kafka.Writer) domain.ProductEventPublisher {
	return &kafkaPublisher{writer: writer}
}

// 1. Define your Event Envelope
type ProductEventEnvelope struct {
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      *domain.Product `json:"data"`
	// CorrelationID string   `json:"correlation_id,omitempty"` // Add later if needed
}

func (k *kafkaPublisher) PublishProductCreated(ctx context.Context, p *domain.Product) error {
	
	// 2. Wrap the product in the envelope
	event := ProductEventEnvelope{
		EventType: "product_created",
		Timestamp: time.Now().UTC(),
		Data:      p,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	traceHeaders := tracing.InjectKafkaHeaders(ctx)

	// 3. Fix the Partition Key (Use Product ID)
	keyString := fmt.Sprintf("product-%d", p.ID)

	msg := kafka.Message{
		Key:     []byte(keyString), // Ensures parallel processing across partitions
		Value:   eventBytes,
		Headers: traceHeaders,
	}

	return k.writer.WriteMessages(ctx, msg)
}