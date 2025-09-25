package events

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
)

type stubWriter struct {
	written []string
	err     error
	closed  bool
}

func (s *stubWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if s.err != nil {
		return s.err
	}
	for _, m := range msgs {
		s.written = append(s.written, string(m.Value))
	}
	return nil
}

func (s *stubWriter) Close() error {
	s.closed = true
	return nil
}

func TestProducerPublishSuccess(t *testing.T) {
	writer := &stubWriter{}
	producer := &Producer{writer: writer}

	payload := map[string]string{"event": "user.created"}
	if err := producer.Publish(context.Background(), "key", payload); err != nil {
		t.Fatalf("Publish error: %v", err)
	}
	if len(writer.written) != 1 {
		t.Fatalf("expected one message to be written")
	}
	if writer.written[0] == "" {
		t.Fatalf("expected serialized payload")
	}
}

func TestProducerPublishMarshalError(t *testing.T) {
	writer := &stubWriter{}
	producer := &Producer{writer: writer}

	if err := producer.Publish(context.Background(), "key", func() {}); err == nil {
		t.Fatalf("expected marshal error")
	}
}

func TestProducerClose(t *testing.T) {
	writer := &stubWriter{}
	producer := &Producer{writer: writer}

	if err := producer.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if !writer.closed {
		t.Fatalf("expected writer to be closed")
	}
}

func TestProducerWriterError(t *testing.T) {
	writer := &stubWriter{err: errors.New("write failure")}
	producer := &Producer{writer: writer}
	if err := producer.Publish(context.Background(), "key", map[string]string{"test": "data"}); err == nil {
		t.Fatalf("expected write error")
	}
}

func TestNewProducerConfiguresWriter(t *testing.T) {
	producer := NewProducer([]string{"localhost:9092"}, "users")
	if producer.writer == nil {
		t.Fatalf("expected writer to be initialized")
	}
	if err := producer.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
}
