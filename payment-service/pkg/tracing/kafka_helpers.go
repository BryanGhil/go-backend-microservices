package tracing

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

// InjectKafkaHeaders takes your Go context and turns it into Kafka headers
func InjectKafkaHeaders(ctx context.Context) []kafka.Header {
	var headers []kafka.Header
	carrier := KafkaCarrier{Headers: &headers}
	
	// Stamp the Trace ID onto the headers
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	
	return headers
}

// ExtractKafkaContext reads Kafka headers and gives you back a linked Go context
func ExtractKafkaContext(ctx context.Context, headers []kafka.Header) context.Context {
	carrier := KafkaCarrier{Headers: &headers}
	
	// Pull the Trace ID out and attach it to the new context
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}