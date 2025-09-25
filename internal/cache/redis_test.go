package cache

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestNewClientConfiguration(t *testing.T) {
	client := NewClient("localhost:6379")
	opts := client.Options()
	if opts.Addr != "localhost:6379" {
		t.Fatalf("Addr = %s, want localhost:6379", opts.Addr)
	}
	if opts.DialTimeout != 5*time.Second {
		t.Fatalf("DialTimeout = %v", opts.DialTimeout)
	}
	if opts.PoolSize != 10 {
		t.Fatalf("PoolSize = %d, want 10", opts.PoolSize)
	}
	_ = client.Close()
}

func TestPingSuccess(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	client := redis.NewClient(&redis.Options{
		Protocol: 2,
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return clientConn, nil
		},
	})
	defer client.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer serverConn.Close()
		buf := make([]byte, 256)
		for {
			n, err := serverConn.Read(buf)
			if err != nil {
				return
			}
			payload := strings.ToLower(string(buf[:n]))
			switch {
			case strings.Contains(payload, "hello"):
				_, _ = serverConn.Write([]byte("%2\r\n+server\r\n+redis\r\n+proto\r\n:2\r\n"))
			case strings.Contains(payload, "client"):
				count := strings.Count(payload, "client")
				for i := 0; i < count; i++ {
					_, _ = serverConn.Write([]byte("+OK\r\n"))
				}
			case strings.Contains(payload, "ping"):
				_, _ = serverConn.Write([]byte("+PONG\r\n"))
				return
			default:
				return
			}
		}
	}()

	if err := Ping(context.Background(), client); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	<-done
}

func TestPingError(t *testing.T) {
	wantErr := errors.New("dial error")
	client := redis.NewClient(&redis.Options{
		Protocol: 2,
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return nil, wantErr
		},
	})
	defer client.Close()

	if err := Ping(context.Background(), client); !errors.Is(err, wantErr) {
		t.Fatalf("Ping() error = %v, want %v", err, wantErr)
	}
}
