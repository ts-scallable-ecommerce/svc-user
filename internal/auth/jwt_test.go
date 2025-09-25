package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func generateKeyPair(t *testing.T) ([]byte, []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	privateBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	publicBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)})

	return privateBytes, publicBytes
}

func TestTokenIssuerLifecycle(t *testing.T) {
	priv, pub := generateKeyPair(t)

	issuer, err := NewTokenIssuer(priv, pub, "svc-user", []string{"api"})
	if err != nil {
		t.Fatalf("NewTokenIssuer returned error: %v", err)
	}

	token, err := issuer.GenerateAccessToken("user-123", map[string]any{"role": "admin"})
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}
	if token == "" {
		t.Fatalf("expected token to be issued")
	}

	claims, err := issuer.ParseAndValidate(token)
	if err != nil {
		t.Fatalf("ParseAndValidate returned error: %v", err)
	}
	if claims["sub"] != "user-123" {
		t.Fatalf("unexpected subject: %v", claims["sub"])
	}
	if claims["role"] != "admin" {
		t.Fatalf("expected role claim to be preserved")
	}

	if issuer.AccessTokenTTL() <= 0 {
		t.Fatalf("expected positive access token TTL")
	}
	if issuer.RefreshTokenTTL() <= issuer.AccessTokenTTL() {
		t.Fatalf("expected refresh TTL to exceed access TTL")
	}
}

func TestParseAndValidateRejectsUnexpectedSigningMethod(t *testing.T) {
	priv, pub := generateKeyPair(t)
	issuer, err := NewTokenIssuer(priv, pub, "svc-user", []string{"api"})
	if err != nil {
		t.Fatalf("NewTokenIssuer returned error: %v", err)
	}

	hmacToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "user"}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("failed to sign HMAC token: %v", err)
	}

	if _, err := issuer.ParseAndValidate(hmacToken); err == nil {
		t.Fatalf("expected ParseAndValidate to reject unexpected signing method")
	}

	if _, err := issuer.ParseAndValidate("not a token"); err == nil {
		t.Fatalf("expected parse error for malformed token")
	}
}

func TestNewTokenIssuerErrors(t *testing.T) {
	if _, err := NewTokenIssuer([]byte("not a key"), []byte("pub"), "iss", nil); err == nil {
		t.Fatalf("expected error for invalid private key")
	}

	priv, _ := generateKeyPair(t)
	if _, err := NewTokenIssuer(priv, []byte("bad"), "iss", nil); err == nil {
		t.Fatalf("expected error for invalid public key")
	}
}

func TestLoadIssuerFromFiles(t *testing.T) {
	priv, pub := generateKeyPair(t)
	dir := t.TempDir()
	privPath := filepath.Join(dir, "priv.pem")
	pubPath := filepath.Join(dir, "pub.pem")

	if err := os.WriteFile(privPath, priv, 0o600); err != nil {
		t.Fatalf("failed to write private key: %v", err)
	}
	if err := os.WriteFile(pubPath, pub, 0o600); err != nil {
		t.Fatalf("failed to write public key: %v", err)
	}

	issuer, err := LoadIssuerFromFiles(privPath, pubPath, "svc", []string{"aud"})
	if err != nil {
		t.Fatalf("LoadIssuerFromFiles returned error: %v", err)
	}
	if issuer == nil {
		t.Fatalf("expected issuer instance")
	}

	if _, err := LoadIssuerFromFiles("missing", pubPath, "svc", nil); err == nil {
		t.Fatalf("expected error when private key missing")
	}

	if _, err := LoadIssuerFromFiles(privPath, "missing", "svc", nil); err == nil {
		t.Fatalf("expected error when public key missing")
	}
}
