package tracing

import "github.com/segmentio/kafka-go"

// KafkaCarrier adapts kafka.Header to the OpenTelemetry TextMapCarrier interface.
type KafkaCarrier struct {
	Headers *[]kafka.Header
}

// Get returns the value associated with the passed key.
func (c KafkaCarrier) Get(key string) string {
	for _, h := range *c.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set stores the key-value pair.
func (c KafkaCarrier) Set(key string, value string) {
	*c.Headers = append(*c.Headers, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

// Keys lists the keys stored in this carrier.
func (c KafkaCarrier) Keys() []string {
	keys := make([]string, len(*c.Headers))
	for i, h := range *c.Headers {
		keys[i] = h.Key
	}
	return keys
}