package grpc

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestServerLifecycle(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate port: %v", err)
	}
	addr := lis.Addr().String()
	lis.Close()

	srv := NewServer(addr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve()
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(25 * time.Millisecond)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	srv.GracefulStop(ctx)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("gRPC server did not stop in time")
	}

	if srv.Underlying() == nil {
		t.Fatalf("expected underlying server to be exposed")
	}
}

func TestServerServeListenError(t *testing.T) {
	srv := NewServer("127.0.0.1:-1")
	if err := srv.Serve(); err == nil {
		t.Fatalf("expected listen error")
	}
}

func TestGracefulStopWithTimeout(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate port: %v", err)
	}
	addr := lis.Addr().String()
	lis.Close()

	srv := NewServer(addr)
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve() }()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(25 * time.Millisecond)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	cancel()
	srv.GracefulStop(ctx)

	select {
	case <-errCh:
	case <-time.After(time.Second):
		t.Fatalf("server did not stop after timeout")
	}
}
