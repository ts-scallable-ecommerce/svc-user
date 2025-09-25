package handlers

import (
        "net/http"
        "testing"

        "github.com/gofiber/fiber/v2"
)

func TestRegisterHealthRoutes(t *testing.T) {
        app := fiber.New()
        RegisterHealthRoutes(app)

        tests := []struct {
                path string
        }{
                {"/healthz"},
                {"/livez"},
        }

        for _, tc := range tests {
                req, _ := http.NewRequest(http.MethodGet, tc.path, nil)
                resp, err := app.Test(req)
                if err != nil {
                        t.Fatalf("app.Test(%s) returned error: %v", tc.path, err)
                }
                if resp.StatusCode != http.StatusOK {
                        t.Fatalf("expected status 200, got %d", resp.StatusCode)
                }
        }
}
