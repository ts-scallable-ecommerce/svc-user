package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/response"
)

const (
	userIDContextKey = "user_id"
	tokenContextKey  = "auth_token"
)

// Authenticated parses the Authorization header and injects the authenticated subject into the context.
func Authenticated(issuer *auth.TokenIssuer, blacklist auth.TokenBlacklist) fiber.Handler {
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

		if blacklist != nil {
			revoked, err := blacklist.IsBlacklisted(c.Context(), token)
			if err != nil {
				return response.InternalError(c, "failed to validate token")
			}
			if revoked {
				return response.Unauthorized(c, "token revoked")
			}
		}

		sub, err := issuer.SubjectFromToken(token)
		if err != nil {
			return response.Unauthorized(c, "invalid or expired token")
		}

		c.Locals(userIDContextKey, sub)
		c.Locals(tokenContextKey, token)
		return c.Next()
	}
}

// UserID extracts the authenticated subject from the context.
func UserID(c *fiber.Ctx) string {
	userID, _ := c.Locals(userIDContextKey).(string)
	return userID
}

// Token extracts the bearer token from the context.
func Token(c *fiber.Ctx) string {
	token, _ := c.Locals(tokenContextKey).(string)
	return token
}
