package middleware

import (
        "bytes"
        "encoding/json"
        "time"

        "github.com/gofiber/fiber/v2"
        "log/slog"

        "github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/response"
)

// Logger provides structured request/response logging using slog.
func Logger(log *slog.Logger) fiber.Handler {
        return func(c *fiber.Ctx) error {
                start := time.Now()
                requestID, _ := c.Locals("request_id").(string)

                log.InfoContext(c.Context(), "http request",
                        slog.String("method", c.Method()),
                        slog.String("path", c.Path()),
                        slog.String("ip", c.IP()),
                        slog.String("request_id", requestID),
                )

                if err := c.Next(); err != nil {
                        // Let Fiber handle the error so the configured error handler emits the response.
                        return err
                }

                latency := time.Since(start)
                status := c.Response().StatusCode()
                var respPayload any

                if stored := c.Locals(response.ContextKey); stored != nil {
                        respPayload = stored
                } else if body := c.Response().Body(); len(body) > 0 {
                        // Attempt to decode JSON into a generic structure for readability.
                        var parsed any
                        if err := json.Unmarshal(body, &parsed); err == nil {
                                respPayload = parsed
                        } else {
                                // fallback to raw body to avoid losing response information
                                respPayload = string(bytes.Clone(body))
                        }
                }

                log.InfoContext(c.Context(), "http response",
                        slog.Int("status", status),
                        slog.Duration("latency", latency),
                        slog.String("request_id", requestID),
                        slog.Any("response", respPayload),
                )

                return nil
        }
}

