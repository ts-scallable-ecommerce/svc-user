package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/response"
)

const userIDContextKey = "user_id"

// Authenticated parses the Authorization header and injects the authenticated subject into the context.
func Authenticated(issuer *auth.TokenIssuer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" {
			return response.Unauthorized(c, "missing authorization header")
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return response.Unauthorized(c, "invalid authorization header")
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		if token == "" {
			return response.Unauthorized(c, "missing bearer token")
		}

		sub, err := issuer.SubjectFromToken(token)
		if err != nil {
			return response.Unauthorized(c, "invalid or expired token")
		}

		c.Locals(userIDContextKey, sub)
		return c.Next()
	}
}

// UserID extracts the authenticated subject from the context.
func UserID(c *fiber.Ctx) string {
	userID, _ := c.Locals(userIDContextKey).(string)
	return userID
}
