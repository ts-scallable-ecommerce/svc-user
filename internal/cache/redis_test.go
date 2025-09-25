package cache

import (
        "context"
        "testing"
        "time"

        miniredis "github.com/alicebob/miniredis/v2"
)

func TestRedisClientPing(t *testing.T) {
        mr, err := miniredis.Run()
        if err != nil {
                t.Fatalf("failed to start miniredis: %v", err)
        }
        defer func() {
                if mr != nil {
                        mr.Close()
                }
        }()

        client := NewClient(mr.Addr())
        ctx, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel()

        if err := Ping(ctx, client); err != nil {
                t.Fatalf("Ping returned error: %v", err)
        }

        mr.Close()
        mr = nil
        if err := Ping(ctx, client); err == nil {
                t.Fatalf("expected ping error after server closed")
        }
}
