package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklist provides token revocation operations.
type TokenBlacklist interface {
	Revoke(ctx context.Context, token string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

// RedisTokenBlacklist stores blacklisted tokens in Redis.
type RedisTokenBlacklist struct {
	client *redis.Client
	prefix string
}

// NewRedisTokenBlacklist constructs a Redis-backed blacklist implementation.
func NewRedisTokenBlacklist(client *redis.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{client: client, prefix: "auth:blacklist"}
}

// Revoke stores the token in Redis with a TTL matching the token's expiration.
func (b *RedisTokenBlacklist) Revoke(ctx context.Context, token string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Second
	}
	return b.client.Set(ctx, b.key(token), "revoked", ttl).Err()
}

// IsBlacklisted checks whether the token exists in the blacklist.
func (b *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	res, err := b.client.Exists(ctx, b.key(token)).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

func (b *RedisTokenBlacklist) key(token string) string {
	sum := sha256.Sum256([]byte(token))
	return b.prefix + ":" + hex.EncodeToString(sum[:])
}
