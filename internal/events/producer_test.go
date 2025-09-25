package events

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
)

type stubWriter struct {
	messages []kafka.Message
	err      error
	closed   bool
}

func (s *stubWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if s.err != nil {
		return s.err
	}
	s.messages = append(s.messages, msgs...)
	return nil
}

func (s *stubWriter) Close() error {
	s.closed = true
	return nil
}

func TestProducerPublishesMessages(t *testing.T) {
	writer := &stubWriter{}
	producer := newProducer(writer)

	err := producer.Publish(context.Background(), "user", map[string]string{"event": "UserRegistered"})
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	if len(writer.messages) != 1 {
		t.Fatalf("expected message to be written")
	}
	if string(writer.messages[0].Key) != "user" {
		t.Fatalf("unexpected key: %s", writer.messages[0].Key)
	}

	if err := producer.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if !writer.closed {
		t.Fatalf("expected writer to be closed")
	}
}

func TestProducerPublishMarshallingError(t *testing.T) {
	writer := &stubWriter{}
	producer := newProducer(writer)

	if err := producer.Publish(context.Background(), "user", func() {}); err == nil {
		t.Fatalf("expected marshalling error")
	}
}

func TestProducerPublishWriteError(t *testing.T) {
	writer := &stubWriter{err: errors.New("write failed")}
	producer := newProducer(writer)

	if err := producer.Publish(context.Background(), "user", map[string]string{"event": "UserRegistered"}); err == nil {
		t.Fatalf("expected writer error to propagate")
	}
}

func TestNewProducerCreatesKafkaWriter(t *testing.T) {
	producer := NewProducer([]string{"localhost:9092"}, "events")
	if producer == nil {
		t.Fatalf("expected producer instance")
	}
	if err := producer.Close(); err != nil {
		t.Fatalf("expected Close to succeed: %v", err)
	}
}
