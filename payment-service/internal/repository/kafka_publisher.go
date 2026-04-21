package repository

import (
	"context"
	"ecommerce/payment-service/internal/domain"
	"ecommerce/payment-service/pkg/tracing"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type kafkaPub struct{ writer *kafka.Writer }

func NewKafkaPublisher(w *kafka.Writer) domain.KafkaPublisher { return &kafkaPub{writer: w} }

func (k *kafkaPub) PublishEvent(ctx context.Context, topic string, eventType string, event domain.SagaEvent) error {
	bytes, _ := json.Marshal(event)

	traceHeaders := tracing.InjectKafkaHeaders(ctx)
	
	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(eventType),
		Value: bytes,
		Headers: traceHeaders,
	})
}