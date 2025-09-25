package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRegisterHealthRoutes(t *testing.T) {
	app := fiber.New()
	RegisterHealthRoutes(app)

	req := httptest.NewRequest("GET", "/healthz", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("unexpected health status: %v", body)
	}

	liveReq := httptest.NewRequest("GET", "/livez", nil)
	liveResp, err := app.Test(liveReq)
	if err != nil {
		t.Fatalf("livez request failed: %v", err)
	}
	if liveResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", liveResp.StatusCode)
	}
	body = map[string]string{}
	if err := json.NewDecoder(liveResp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode live response: %v", err)
	}
	if body["status"] != "alive" {
		t.Fatalf("unexpected live status: %v", body)
	}
}
