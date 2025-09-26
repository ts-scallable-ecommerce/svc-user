package http

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log/slog"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/handlers"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/middleware"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/response"
)

// Server wraps the Fiber app and configuration.
type Server struct {
	app *fiber.App
	cfg *config.Config
}

// NewServer configures the HTTP server with middlewares and routes.
func NewServer(cfg *config.Config, log *slog.Logger, issuer *auth.TokenIssuer, userHandler *handlers.UserHandler) (*Server, error) {
	app := fiber.New(fiber.Config{
		Prefork:               false,
		DisableStartupMessage: true,
		ReadTimeout:           cfg.ReadTimeout,
		WriteTimeout:          cfg.WriteTimeout,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				return response.JSON(c, fiberErr.Code, fiberErr.Message, nil)
			}
			return response.InternalError(c, err.Error())
		},
	})

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger(log))

	handlers.RegisterHealthRoutes(app)

	api := app.Group("/api/v1")
	handlers.RegisterUserRoutes(api, userHandler, middleware.Authenticated(issuer))

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
