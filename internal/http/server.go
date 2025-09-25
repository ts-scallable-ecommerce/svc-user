package http

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/handlers"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/middleware"
)

// Server wraps the Fiber app and configuration.
type Server struct {
	app *fiber.App
	cfg *config.Config
}

// NewServer configures the HTTP server with middlewares and routes.
func NewServer(cfg *config.Config) (*Server, error) {
	app := fiber.New(fiber.Config{
		Prefork:               false,
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(middleware.RequestID())

	handlers.RegisterHealthRoutes(app)

	return &Server{app: app, cfg: cfg}, nil
}

// Start begins listening on the configured HTTP address.
func (s *Server) Start() error {
	return s.app.Listen(s.cfg.HTTPAddr)
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	return nil
}
