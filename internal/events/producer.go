package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// Producer publishes domain events to Kafka topics using the outbox pattern.
type messageWriter interface {
	WriteMessages(context.Context, ...kafka.Message) error
	Close() error
}

// Producer publishes domain events to Kafka topics using the outbox pattern.
type Producer struct {
	writer messageWriter
}

// NewProducer configures the Kafka writer.
func NewProducer(brokers []string, topic string) *Producer {
	return newProducer(&kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
}

func newProducer(writer messageWriter) *Producer {
	return &Producer{writer: writer}
}

// Publish emits an event payload.
func (p *Producer) Publish(ctx context.Context, key string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: body,
	})
}

// Close releases resources held by the producer.
func (p *Producer) Close() error {
	return p.writer.Close()
}
