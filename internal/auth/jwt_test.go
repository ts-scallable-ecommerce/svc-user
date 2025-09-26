package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
)

func TestTokenIssuer(t *testing.T) {
	privPEM, pubPEM := generateKeyPair(t)

	issuer, err := auth.NewTokenIssuer(privPEM, pubPEM, "svc-user", []string{"test"})
	if err != nil {
		t.Fatalf("new token issuer: %v", err)
	}

	access, err := issuer.GenerateAccessToken("user-123", map[string]any{"role": "admin"})
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}
	if access == "" {
		t.Fatal("access token should not be empty")
	}

	refresh, err := issuer.GenerateRefreshToken("user-123", nil)
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}
	if refresh == "" {
		t.Fatal("refresh token should not be empty")
	}

	subject, err := issuer.SubjectFromToken(access)
	if err != nil {
		t.Fatalf("parse subject: %v", err)
	}
	if subject != "user-123" {
		t.Fatalf("expected subject user-123 got %s", subject)
	}
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
