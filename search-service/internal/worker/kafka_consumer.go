package worker

import (
	"context"
	"ecommerce/search-service/internal/domain"
	"ecommerce/search-service/pkg/tracing"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	repo   domain.SearchRepository
}

func NewKafkaConsumer(reader *kafka.Reader, repo domain.SearchRepository) *KafkaConsumer {
	return &KafkaConsumer{reader: reader, repo: repo}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Search Service: Listening for Kafka events...")
	tracer := otel.Tracer("search-consumer")

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		func() {
			linkedCtx := tracing.ExtractKafkaContext(context.Background(), m.Headers)
			spanCtx, span := tracer.Start(linkedCtx, "ConsumeKafkaMessage: "+string(m.Key))
			defer span.End()

			// 1. Unmarshal into the Envelope first
			var envelope domain.ProductEventEnvelope
			if err := json.Unmarshal(m.Value, &envelope); err != nil {
				log.Printf("Failed to unmarshal event envelope: %v", err)
				return
			}

			// 2. Check the Event Type
			if envelope.EventType == "product_created" {
				// 3. Index the actual product data inside the envelope
				err = c.repo.IndexProduct(spanCtx, envelope.Data)
				if err != nil {
					log.Printf("Failed to index product %d in ES: %v", envelope.Data.ID, err)
				} else {
					log.Printf("Successfully indexed product %d into Elasticsearch!", envelope.Data.ID)
				}
			}
		}()
	}
}