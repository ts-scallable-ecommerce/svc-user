package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRegisterHealthRoutes(t *testing.T) {
	app := fiber.New()
	RegisterHealthRoutes(app)

	tests := []struct {
		path     string
		expected string
	}{
		{"/healthz", "ok"},
		{"/livez", "alive"},
	}

	for _, tc := range tests {
		resp, err := app.Test(httptest.NewRequest("GET", tc.path, nil), -1)
		if err != nil {
			t.Fatalf("app.Test(%s) error = %v", tc.path, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status for %s = %d, want %d", tc.path, resp.StatusCode, fiber.StatusOK)
		}
	}
}
