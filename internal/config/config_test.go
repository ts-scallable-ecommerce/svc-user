package config

import (
        "testing"
        "time"
)

func TestLoadConfigFromEnvironment(t *testing.T) {
        t.Setenv("DB_DSN", "postgres://user:pass@localhost/db")
        t.Setenv("HTTP_ADDR", ":9090")
        t.Setenv("HTTP_READ_TIMEOUT_SECONDS", "30")
        t.Setenv("HTTP_WRITE_TIMEOUT_SECONDS", "45")
        t.Setenv("HTTP_GRACEFUL_TIMEOUT_SECONDS", "5")
        t.Setenv("REDIS_ADDR", "localhost:6380")
        t.Setenv("JWT_PRIVATE_KEY_PATH", "/tmp/priv.pem")
        t.Setenv("JWT_PUBLIC_KEY_PATH", "/tmp/pub.pem")

        cfg, err := Load()
        if err != nil {
                t.Fatalf("Load returned error: %v", err)
        }

        if cfg.HTTPAddr != ":9090" {
                t.Fatalf("unexpected HTTPAddr: %s", cfg.HTTPAddr)
        }
        if cfg.ReadTimeout != 30*time.Second {
                t.Fatalf("expected read timeout 30s, got %s", cfg.ReadTimeout)
        }
        if cfg.WriteTimeout != 45*time.Second {
                t.Fatalf("expected write timeout 45s, got %s", cfg.WriteTimeout)
        }
        if cfg.GracefulTimeout != 5*time.Second {
                t.Fatalf("expected graceful timeout 5s, got %s", cfg.GracefulTimeout)
        }
        if cfg.RedisAddr != "localhost:6380" {
                t.Fatalf("unexpected redis address: %s", cfg.RedisAddr)
        }
        if cfg.JWTPrivateKeyPath != "/tmp/priv.pem" || cfg.JWTPublicKeyPath != "/tmp/pub.pem" {
                t.Fatalf("expected JWT key paths to be preserved")
        }
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
        t.Setenv("DB_DSN", "")
        if _, err := Load(); err == nil {
                t.Fatalf("expected error when DB_DSN is missing")
        }
}
