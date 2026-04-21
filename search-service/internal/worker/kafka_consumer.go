package worker

import (
	"context"
	"ecommerce/search-service/pkg/tracing"
	"ecommerce/search-service/internal/domain"
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
		// This blocks until a new message arrives
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		func() {
			linkedCtx := tracing.ExtractKafkaContext(context.Background(), m.Headers)

			// 2. Start the span using the linked context
			spanCtx, span := tracer.Start(linkedCtx, "ConsumeKafkaMessage: "+string(m.Key))
			defer span.End() // Safely closes when the function finishes

			if string(m.Key) == "product-created" {
				var product domain.Product
				if err := json.Unmarshal(m.Value, &product); err == nil {
					// Save it to Elasticsearch!
					err = c.repo.IndexProduct(spanCtx, &product)
					if err != nil {
						log.Printf("Failed to index product in ES: %v", err)
					} else {
						log.Printf("Successfully indexed product %d into Elasticsearch!", product.ID)
					}
				}
			}

		}()

	}
}
