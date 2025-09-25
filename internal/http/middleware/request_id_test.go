package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestIDGeneratesNewID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	var header string
	var local any
	app.Get("/", func(c *fiber.Ctx) error {
		header = c.GetRespHeader("X-Request-ID")
		local = c.Locals("request_id")
		return nil
	})

	if _, err := app.Test(httptest.NewRequest("GET", "/", nil)); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if header == "" {
		t.Fatalf("expected response header to be set")
	}
	if _, ok := local.(string); !ok {
		t.Fatalf("expected request_id local to be string, got %T", local)
	}
}

func TestRequestIDPreservesExisting(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	var header string
	app.Get("/", func(c *fiber.Ctx) error {
		header = c.GetRespHeader("X-Request-ID")
		return nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "existing")
	if _, err := app.Test(req); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if header != "existing" {
		t.Fatalf("expected existing header to be preserved, got %s", header)
	}
}
