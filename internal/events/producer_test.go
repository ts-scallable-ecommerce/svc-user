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
	return s.err
}

func TestProducerPublish(t *testing.T) {
	w := &stubWriter{}
	producer := NewProducerWithWriter(w)

	payload := map[string]string{"event": "user.created"}
	if err := producer.Publish(context.Background(), "key", payload); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if len(w.messages) != 1 {
		t.Fatalf("expected one message written, got %d", len(w.messages))
	}
	if string(w.messages[0].Key) != "key" {
		t.Fatalf("unexpected message key: %s", w.messages[0].Key)
	}
	if err := producer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !w.closed {
		t.Fatal("expected writer to be closed")
	}
}

func TestNewProducerCreatesKafkaWriter(t *testing.T) {
	producer := NewProducer([]string{"localhost:9092"}, "users")
	if producer == nil {
		t.Fatal("expected producer instance")
	}
	if err := producer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestProducerPublishError(t *testing.T) {
	want := errors.New("write failed")
	w := &stubWriter{err: want}
	producer := NewProducerWithWriter(w)

	if err := producer.Publish(context.Background(), "key", map[string]string{}); !errors.Is(err, want) {
		t.Fatalf("Publish() error = %v, want %v", err, want)
	}
	if err := producer.Close(); !errors.Is(err, want) {
		t.Fatalf("Close() error = %v, want %v", err, want)
	}
}

func TestProducerPublishMarshalError(t *testing.T) {
	w := &stubWriter{}
	producer := NewProducerWithWriter(w)

	type invalid struct{}
	ch := make(chan int)
	payload := map[string]interface{}{"invalid": invalid{}, "chan": ch}

	if err := producer.Publish(context.Background(), "key", payload); err == nil {
		t.Fatal("expected marshal error")
	}
}
