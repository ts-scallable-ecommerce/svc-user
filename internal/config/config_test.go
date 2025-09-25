package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadSuccessWithOverrides(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://localhost:5432/db")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("HTTP_READ_TIMEOUT_SECONDS", "20")
	t.Setenv("HTTP_WRITE_TIMEOUT_SECONDS", "25")
	t.Setenv("HTTP_GRACEFUL_TIMEOUT_SECONDS", "30")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "/tmp/private.pem")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "/tmp/public.pem")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Errorf("HTTPAddr = %s, want %s", cfg.HTTPAddr, ":9090")
	}
	if cfg.RedisAddr != "redis:6379" {
		t.Errorf("RedisAddr = %s, want %s", cfg.RedisAddr, "redis:6379")
	}
	if cfg.ReadTimeout != 20*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", cfg.ReadTimeout, 20*time.Second)
	}
	if cfg.WriteTimeout != 25*time.Second {
		t.Errorf("WriteTimeout = %v, want %v", cfg.WriteTimeout, 25*time.Second)
	}
	if cfg.GracefulTimeout != 30*time.Second {
		t.Errorf("GracefulTimeout = %v, want %v", cfg.GracefulTimeout, 30*time.Second)
	}
	if cfg.JWTPrivateKeyPath != "/tmp/private.pem" {
		t.Errorf("JWTPrivateKeyPath = %s, want /tmp/private.pem", cfg.JWTPrivateKeyPath)
	}
	if cfg.JWTPublicKeyPath != "/tmp/public.pem" {
		t.Errorf("JWTPublicKeyPath = %s, want /tmp/public.pem", cfg.JWTPublicKeyPath)
	}
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DB_DSN")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected error when DB_DSN missing")
	}
}

func TestLoadFallsBackToDefaults(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://localhost:5432/db")
	t.Setenv("HTTP_READ_TIMEOUT_SECONDS", "invalid")
	t.Setenv("HTTP_WRITE_TIMEOUT_SECONDS", "")
	t.Setenv("HTTP_GRACEFUL_TIMEOUT_SECONDS", "invalid")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr default mismatch: %s", cfg.HTTPAddr)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("ReadTimeout fallback mismatch: %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 15*time.Second {
		t.Errorf("WriteTimeout fallback mismatch: %v", cfg.WriteTimeout)
	}
	if cfg.GracefulTimeout != 10*time.Second {
		t.Errorf("GracefulTimeout fallback mismatch: %v", cfg.GracefulTimeout)
	}
}
