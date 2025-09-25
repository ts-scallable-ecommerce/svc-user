package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr          string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	GracefulTimeout   time.Duration
	DatabaseURL       string
	RedisAddr         string
	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:       os.Getenv("DB_DSN"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		JWTPrivateKeyPath: os.Getenv("JWT_PRIVATE_KEY_PATH"),
		JWTPublicKeyPath:  os.Getenv("JWT_PUBLIC_KEY_PATH"),
		ReadTimeout:       getDurationEnv("HTTP_READ_TIMEOUT_SECONDS", 15*time.Second),
		WriteTimeout:      getDurationEnv("HTTP_WRITE_TIMEOUT_SECONDS", 15*time.Second),
		GracefulTimeout:   getDurationEnv("HTTP_GRACEFUL_TIMEOUT_SECONDS", 10*time.Second),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DB_DSN environment variable must be set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		seconds, err := strconv.Atoi(val)
		if err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return fallback
}
