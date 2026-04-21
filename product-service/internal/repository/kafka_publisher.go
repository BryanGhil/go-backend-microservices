package repository

import (
	"context"
	"ecommerce/product-service/internal/domain"
	"ecommerce/product-service/pkg/tracing"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type kafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(writer *kafka.Writer) domain.ProductEventPublisher {
	return &kafkaPublisher{writer: writer}
}

func (k *kafkaPublisher) PublishProductCreated(ctx context.Context, p *domain.Product) error {
	// 1. Convert the Go struct to JSON
	eventBytes, err := json.Marshal(p)
	if err != nil {
		return err
	}

	traceHeaders := tracing.InjectKafkaHeaders(ctx)

	// 2. Publish to the "product-events" topic
	msg := kafka.Message{
		Key:   []byte("product-created"), // Helps categorize the event
		Value: eventBytes,
		Headers: traceHeaders,
	}

	return k.writer.WriteMessages(ctx, msg)
}