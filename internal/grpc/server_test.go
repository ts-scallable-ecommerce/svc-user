package grpc

import (
	"context"
	"testing"
	"time"
)

func TestNewServerExposesUnderlying(t *testing.T) {
	srv := NewServer("127.0.0.1:0")
	if srv.Underlying() == nil {
		t.Fatalf("expected underlying server")
	}
}

func TestServerServeAndGracefulStop(t *testing.T) {
	srv := NewServer("127.0.0.1:0")
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve()
	}()

	select {
	case <-time.After(200 * time.Millisecond):
	case err := <-errCh:
		t.Fatalf("serve exited early: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	srv.GracefulStop(ctx)

	if err := <-errCh; err != nil {
		t.Fatalf("serve returned error: %v", err)
	}
}

func TestGracefulStopCancelsContext(t *testing.T) {
	srv := NewServer("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	srv.GracefulStop(ctx)
}

func TestServeListenError(t *testing.T) {
	srv := NewServer(":::-invalid")
	if err := srv.Serve(); err == nil {
		t.Fatalf("expected listen error")
	}
}
