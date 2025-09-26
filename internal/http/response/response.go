package response

import "github.com/gofiber/fiber/v2"

// ContextKey represents the fiber context key storing the base response payload.
const ContextKey = "base_response"

// Base represents the canonical envelope returned by every HTTP handler.
type Base struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// JSON writes the base response and records it for logging middleware.
func JSON(c *fiber.Ctx, status int, message string, data any) error {
	payload := Base{Status: status, Message: message}
	if data != nil {
		payload.Data = data
	}
	c.Locals(ContextKey, payload)
	return c.Status(status).JSON(payload)
}

// OK writes a 200 response.
func OK(c *fiber.Ctx, message string, data any) error {
	return JSON(c, fiber.StatusOK, message, data)
}

// Created writes a 201 response.
func Created(c *fiber.Ctx, message string, data any) error {
	return JSON(c, fiber.StatusCreated, message, data)
}

// Accepted writes a 202 response.
func Accepted(c *fiber.Ctx, message string, data any) error {
	return JSON(c, fiber.StatusAccepted, message, data)
}

// BadRequest writes a 400 response.
func BadRequest(c *fiber.Ctx, message string) error {
	return JSON(c, fiber.StatusBadRequest, message, nil)
}

// Unauthorized writes a 401 response.
func Unauthorized(c *fiber.Ctx, message string) error {
	return JSON(c, fiber.StatusUnauthorized, message, nil)
}

// Forbidden writes a 403 response.
func Forbidden(c *fiber.Ctx, message string) error {
	return JSON(c, fiber.StatusForbidden, message, nil)
}

// NotFound writes a 404 response.
func NotFound(c *fiber.Ctx, message string) error {
	return JSON(c, fiber.StatusNotFound, message, nil)
}

// InternalError writes a 500 response.
func InternalError(c *fiber.Ctx, message string) error {
	return JSON(c, fiber.StatusInternalServerError, message, nil)
}
