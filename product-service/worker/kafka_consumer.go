package worker

import (
	"context"
	"ecommerce/product-service/internal/domain"
	"ecommerce/product-service/pkg/tracing"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type UserEventConsumer struct {
	reader *kafka.Reader
	repo   domain.ProductRepository
}

func NewUserEventConsumer(r *kafka.Reader, repo domain.ProductRepository) *UserEventConsumer {
	return &UserEventConsumer{reader: r, repo: repo}
}

type UserEventEnvelope struct {
	EventType string `json:"event_type"`
	Data      struct {
		SellerID int64  `json:"seller_id"`
		ShopName string `json:"shop_name"`
	} `json:"data"`
}

func (c *UserEventConsumer) Start(ctx context.Context) {
	log.Println("Product Service: Listening for User events...")

	tracer := otel.Tracer("product-consumer")

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			continue
		}

		func() {
			carrier := tracing.KafkaCarrier{Headers: &m.Headers}
			extractedCtx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

			spanCtx, span := tracer.Start(extractedCtx, "ConsumeKafkaMessage")
			defer span.End()

			var envelope UserEventEnvelope
			if err := json.Unmarshal(m.Value, &envelope); err != nil {
				log.Printf("Failed to parse user event JSON: %v", err)
				return 
			}

			if envelope.EventType == "shop_name_updated" {
				log.Printf("Received shop update for Seller %d: New Name = '%s'", envelope.Data.SellerID, envelope.Data.ShopName)

				err := c.repo.UpdateSellerShopName(spanCtx, envelope.Data.SellerID, envelope.Data.ShopName)
				if err != nil {
					log.Printf("Failed to update product DB: %v", err)
				} else {
					log.Printf("Successfully updated all products for Seller %d", envelope.Data.SellerID)
				}
			}

		}()
	}
}
