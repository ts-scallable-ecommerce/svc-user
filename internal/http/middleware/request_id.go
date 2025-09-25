package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID injects a deterministic request ID into the context when one is not provided.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if id := c.Get("X-Request-ID"); id != "" {
			c.Set("X-Request-ID", id)
			return c.Next()
		}
		reqID := uuid.NewString()
		c.Set("X-Request-ID", reqID)
		c.Locals("request_id", reqID)
		return c.Next()
	}
}
