package config

import (
	"testing"
	"time"
)

func TestLoadUsesEnvironmentAndFallbacks(t *testing.T) {
	t.Setenv("DB_DSN", "postgres://user:pass@localhost/db")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("HTTP_READ_TIMEOUT_SECONDS", "30")
	t.Setenv("HTTP_WRITE_TIMEOUT_SECONDS", "45")
	t.Setenv("HTTP_GRACEFUL_TIMEOUT_SECONDS", "20")
	t.Setenv("JWT_PRIVATE_KEY_PATH", "/path/to/private")
	t.Setenv("JWT_PUBLIC_KEY_PATH", "/path/to/public")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("unexpected HTTPAddr: %s", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost/db" {
		t.Fatalf("unexpected DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.ReadTimeout != 30*time.Second || cfg.WriteTimeout != 45*time.Second {
		t.Fatalf("unexpected timeout values: %+v", cfg)
	}
	if cfg.GracefulTimeout != 20*time.Second {
		t.Fatalf("unexpected graceful timeout: %v", cfg.GracefulTimeout)
	}
	if cfg.JWTPrivateKeyPath != "/path/to/private" || cfg.JWTPublicKeyPath != "/path/to/public" {
		t.Fatalf("unexpected key paths: %+v", cfg)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DB_DSN", "")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error when DB_DSN not set")
	}
}

func TestGetEnvFallback(t *testing.T) {
	if val := getEnv("NON_EXISTENT", "fallback"); val != "fallback" {
		t.Fatalf("expected fallback, got %s", val)
	}
	t.Setenv("EXISTING", "value")
	if val := getEnv("EXISTING", "fallback"); val != "value" {
		t.Fatalf("expected environment value, got %s", val)
	}
}

func TestGetDurationEnvFallback(t *testing.T) {
	if val := getDurationEnv("DUR_ENV", 5*time.Second); val != 5*time.Second {
		t.Fatalf("expected fallback duration")
	}
	t.Setenv("DUR_ENV", "10")
	if val := getDurationEnv("DUR_ENV", 5*time.Second); val != 10*time.Second {
		t.Fatalf("expected parsed duration")
	}
	t.Setenv("DUR_ENV", "invalid")
	if val := getDurationEnv("DUR_ENV", 7*time.Second); val != 7*time.Second {
		t.Fatalf("expected fallback for invalid duration")
	}
}
