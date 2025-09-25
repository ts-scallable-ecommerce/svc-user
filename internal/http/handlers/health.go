package handlers

import "github.com/gofiber/fiber/v2"

// RegisterHealthRoutes binds the healthcheck endpoints to the router.
func RegisterHealthRoutes(app *fiber.App) {
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Get("/livez", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "alive",
		})
	})
}
