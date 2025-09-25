package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/config"
)

func TestNewServerRegistersHealthRoutes(t *testing.T) {
	cfg := &config.Config{HTTPAddr: ":0", DatabaseURL: "dsn"}
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	resp, err := srv.app.Test(httptest.NewRequest("GET", "/healthz", nil))
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestServerStartAndStop(t *testing.T) {
	cfg := &config.Config{HTTPAddr: "127.0.0.1:0", DatabaseURL: "dsn"}
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	select {
	case <-time.After(200 * time.Millisecond):
	case err := <-errCh:
		t.Fatalf("server exited early: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("Stop error: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("server returned error: %v", err)
	}
}

func TestServerStopContextCancelled(t *testing.T) {
	cfg := &config.Config{HTTPAddr: ":0", DatabaseURL: "dsn"}
	srv := &Server{app: fiber.New(), cfg: cfg, shutdown: func(context.Context) error {
		return errors.New("shutdown failure")
	}}

	if err := srv.Stop(context.Background()); err == nil {
		t.Fatalf("expected shutdown error")
	}
}
