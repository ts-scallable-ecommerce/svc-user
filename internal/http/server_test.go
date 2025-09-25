package http

import (
        "context"
        "net"
        "net/http"
        "testing"
        "time"

        "github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/config"
)

func TestNewServerRegistersRoutes(t *testing.T) {
        cfg := &config.Config{HTTPAddr: "127.0.0.1:0"}
        srv, err := NewServer(cfg)
        if err != nil {
                t.Fatalf("NewServer returned error: %v", err)
        }

        req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
        resp, err := srv.app.Test(req)
        if err != nil {
                t.Fatalf("app.Test returned error: %v", err)
        }
        if resp.StatusCode != http.StatusOK {
                t.Fatalf("expected /healthz to return 200, got %d", resp.StatusCode)
        }
}

func TestServerStartAndStop(t *testing.T) {
        listener, err := net.Listen("tcp", "127.0.0.1:0")
        if err != nil {
                t.Fatalf("failed to allocate port: %v", err)
        }
        addr := listener.Addr().String()
        listener.Close()

        cfg := &config.Config{HTTPAddr: addr}
        srv, err := NewServer(cfg)
        if err != nil {
                t.Fatalf("NewServer returned error: %v", err)
        }

        errCh := make(chan error, 1)
        go func() {
                errCh <- srv.Start()
        }()

        // Wait for the server to begin listening.
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
        if err := srv.Stop(ctx); err != nil {
                t.Fatalf("Stop returned error: %v", err)
        }

        select {
        case err := <-errCh:
                if err != nil {
                        t.Fatalf("Start returned error: %v", err)
                }
        case <-time.After(2 * time.Second):
                t.Fatalf("server did not stop in time")
        }
}
