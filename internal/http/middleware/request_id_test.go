package middleware

import (
        "net/http"
        "testing"

        "github.com/gofiber/fiber/v2"
)

func TestRequestIDMiddleware(t *testing.T) {
        app := fiber.New()
        app.Use(RequestID())
        app.Get("/", func(c *fiber.Ctx) error {
                if c.Locals("request_id") == nil {
                        return fiber.ErrInternalServerError
                }
                return c.SendStatus(http.StatusOK)
        })

        req, _ := http.NewRequest(http.MethodGet, "/", nil)
        resp, err := app.Test(req)
        if err != nil {
                t.Fatalf("app.Test without header returned error: %v", err)
        }
        if resp.Header.Get("X-Request-ID") == "" {
                t.Fatalf("expected X-Request-ID header to be set")
        }

        reqWithHeader, _ := http.NewRequest(http.MethodGet, "/", nil)
        reqWithHeader.Header.Set("X-Request-ID", "existing")
        resp, err = app.Test(reqWithHeader)
        if err != nil {
                t.Fatalf("app.Test with header returned error: %v", err)
        }
        if resp.Header.Get("X-Request-ID") != "existing" {
                t.Fatalf("expected middleware to preserve existing request ID")
        }
}
