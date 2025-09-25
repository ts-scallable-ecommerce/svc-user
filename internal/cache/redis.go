package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewClient returns a configured Redis client.
func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})
}

// Ping ensures the connection is available.
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
