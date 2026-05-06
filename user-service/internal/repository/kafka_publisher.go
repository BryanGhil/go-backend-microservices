package repository

import (
	"context"
	"ecommerce/user-service/internal/domain"
	"ecommerce/user-service/pkg/tracing"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(writer *kafka.Writer) domain.UserEventPublisher {
	return &KafkaPublisher{writer: writer}
}

type UserEventEnvelope struct {
	EventType string `json:"event_type"`
	Timestamp string `json:"timestamp"` // Good practice to include
	Data      struct {
		SellerID int64  `json:"seller_id"`
		ShopName string `json:"shop_name"`
	} `json:"data"`
}

func (k *KafkaPublisher) PublishShopNameUpdated(ctx context.Context, sellerId int64, shopName string) error {
	event := UserEventEnvelope{
		EventType: "shop_name_updated",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	event.Data.SellerID = sellerId
	event.Data.ShopName = shopName

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal kafka event: %w", err)
	}

	traceHeaders := tracing.InjectKafkaHeaders(ctx)

	keyString := fmt.Sprintf("seller-%d", sellerId)

	msg := kafka.Message{
		Key:     []byte(keyString),
		Value:   eventBytes,
		Headers: traceHeaders,
	}

	return k.writer.WriteMessages(ctx, msg)
}
