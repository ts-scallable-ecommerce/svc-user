package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/response"
)

// RegisterHealthRoutes binds the healthcheck endpoints to the router.
func RegisterHealthRoutes(app *fiber.App) {
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return response.OK(c, "service healthy", fiber.Map{"status": "ok"})
	})

	app.Get("/livez", func(c *fiber.Ctx) error {
		return response.OK(c, "service alive", fiber.Map{"status": "alive"})
	})
}
