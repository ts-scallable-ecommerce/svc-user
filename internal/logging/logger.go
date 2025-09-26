package logging

import (
	"log/slog"
	"os"
	"sync"
)

var (
	once   sync.Once
	logger *slog.Logger
)

// New returns a structured slog.Logger configured for JSON output.
func New() *slog.Logger {
	once.Do(func() {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
		logger = slog.New(handler)
	})
	return logger
}
