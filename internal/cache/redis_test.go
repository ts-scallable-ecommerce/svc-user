package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type fakePinger struct {
	err error
}

func (f *fakePinger) Ping(context.Context) *redis.StatusCmd {
	return redis.NewStatusResult("", f.err)
}

func TestNewClientConfiguresOptions(t *testing.T) {
	client := NewClient("localhost:6379")
	opts := client.Options()
	if opts.Addr != "localhost:6379" {
		t.Fatalf("unexpected addr: %s", opts.Addr)
	}
	if opts.DialTimeout != 5*time.Second || opts.ReadTimeout != 3*time.Second || opts.WriteTimeout != 3*time.Second {
		t.Fatalf("unexpected timeout configuration: %+v", opts)
	}
	if opts.PoolSize != 10 {
		t.Fatalf("unexpected pool size: %d", opts.PoolSize)
	}
}

func TestPingPropagatesError(t *testing.T) {
	err := Ping(context.Background(), &fakePinger{err: context.DeadlineExceeded})
	if err == nil {
		t.Fatalf("expected ping error")
	}
}

func TestPingSuccess(t *testing.T) {
	if err := Ping(context.Background(), &fakePinger{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
