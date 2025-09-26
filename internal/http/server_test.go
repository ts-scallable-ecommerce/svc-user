package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"log/slog"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/http/handlers"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/users"
)

func TestServerRegisterAndProfileRoutes(t *testing.T) {
	issuer := testIssuer(t)
	svc := &stubUserService{
		registerFn: func(ctx context.Context, req users.RegisterRequest) (*users.RegisterResult, error) {
			return &users.RegisterResult{UserID: "user-1", Tokens: users.TokenPair{AccessToken: "access", RefreshToken: "refresh"}}, nil
		},
		authenticateFn: func(ctx context.Context, req users.AuthenticateRequest) (*users.AuthenticateResult, error) {
			return &users.AuthenticateResult{UserID: "user-1", Tokens: users.TokenPair{AccessToken: "access", RefreshToken: "refresh"}}, nil
		},
		getProfileFn: func(ctx context.Context, userID string) (*users.Profile, error) {
			return &users.Profile{ID: userID, Email: "user@example.com", FirstName: "Test", Status: "active"}, nil
		},
		updateProfileFn: func(ctx context.Context, userID string, req users.UpdateProfileRequest) (*users.Profile, error) {
			return &users.Profile{ID: userID, Email: "user@example.com", FirstName: req.FirstName, LastName: req.LastName}, nil
		},
		changePasswordFn: func(ctx context.Context, userID, current, new string) error { return nil },
		assignRoleFn:     func(ctx context.Context, userID, role string) error { return nil },
		permissionsFn:    func(ctx context.Context, userID string) ([]string, error) { return []string{"roles:view"}, nil },
		hasPermissionFn: func(ctx context.Context, userID, permission string) (bool, error) {
			return true, nil
		},
	}

	cfg := &config.Config{HTTPAddr: ":0"}
	srv, err := NewServer(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), issuer, noopBlacklist{}, handlers.NewUserHandler(svc))
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "secretpass", "firstName": "Test"})
	req := httptestNewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	resp, err := srv.app.Test(req)
	if err != nil {
		t.Fatalf("register request: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status 202 got %d", resp.StatusCode)
	}

	token := mustIssueToken(t, issuer, "user-1")
	profileReq := httptestNewRequest(http.MethodGet, "/api/v1/users/me", nil)
	profileReq.Header.Set("Authorization", "Bearer "+token)
	profResp, err := srv.app.Test(profileReq)
	if err != nil {
		t.Fatalf("profile request: %v", err)
	}
	if profResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 got %d", profResp.StatusCode)
	}

	var payload struct {
		Status  int             `json:"status"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(profResp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Status != http.StatusOK || payload.Message == "" {
		t.Fatalf("unexpected base response: %+v", payload)
	}
}

func httptestNewRequest(method, url string, body io.Reader) *http.Request {
	if body == nil {
		body = http.NoBody
	}
	req, _ := http.NewRequest(method, url, body)
	return req
}

func mustIssueToken(t *testing.T, issuer *auth.TokenIssuer, userID string) string {
	t.Helper()
	token, err := issuer.GenerateAccessToken(userID, nil)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return token
}

func testIssuer(t *testing.T) *auth.TokenIssuer {
	t.Helper()
	priv, pub := generateKeyPair(t)
	issuer, err := auth.NewTokenIssuer(priv, pub, "svc-user", []string{"test"})
	if err != nil {
		t.Fatalf("new issuer: %v", err)
	}
	return issuer
}

func generateKeyPair(t *testing.T) ([]byte, []byte) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	return privPEM, pubPEM
}

type stubUserService struct {
	registerFn       func(context.Context, users.RegisterRequest) (*users.RegisterResult, error)
	authenticateFn   func(context.Context, users.AuthenticateRequest) (*users.AuthenticateResult, error)
	getProfileFn     func(context.Context, string) (*users.Profile, error)
	updateProfileFn  func(context.Context, string, users.UpdateProfileRequest) (*users.Profile, error)
	changePasswordFn func(context.Context, string, string, string) error
	logoutFn         func(context.Context, string) error
	assignRoleFn     func(context.Context, string, string) error
	permissionsFn    func(context.Context, string) ([]string, error)
	hasPermissionFn  func(context.Context, string, string) (bool, error)
}

type noopBlacklist struct{}

func (noopBlacklist) Revoke(context.Context, string, time.Duration) error { return nil }
func (noopBlacklist) IsBlacklisted(context.Context, string) (bool, error) { return false, nil }

func (s *stubUserService) Register(ctx context.Context, req users.RegisterRequest) (*users.RegisterResult, error) {
	return s.registerFn(ctx, req)
}

func (s *stubUserService) Authenticate(ctx context.Context, req users.AuthenticateRequest) (*users.AuthenticateResult, error) {
	return s.authenticateFn(ctx, req)
}

func (s *stubUserService) GetProfile(ctx context.Context, userID string) (*users.Profile, error) {
	return s.getProfileFn(ctx, userID)
}

func (s *stubUserService) UpdateProfile(ctx context.Context, userID string, req users.UpdateProfileRequest) (*users.Profile, error) {
	return s.updateProfileFn(ctx, userID, req)
}

func (s *stubUserService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	return s.changePasswordFn(ctx, userID, currentPassword, newPassword)
}

func (s *stubUserService) Logout(ctx context.Context, token string) error {
	if s.logoutFn != nil {
		return s.logoutFn(ctx, token)
	}
	return nil
}

func (s *stubUserService) AssignRole(ctx context.Context, userID, role string) error {
	return s.assignRoleFn(ctx, userID, role)
}

func (s *stubUserService) Permissions(ctx context.Context, userID string) ([]string, error) {
	return s.permissionsFn(ctx, userID)
}

func (s *stubUserService) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	return s.hasPermissionFn(ctx, userID, permission)
}
