package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateKeyPair(t *testing.T) ([]byte, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	priv := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pub := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)})
	return priv, pub
}

func TestTokenIssuer_GenerateAndParse(t *testing.T) {
	priv, pub := generateKeyPair(t)
	issuer, err := NewTokenIssuer(priv, pub, "svc", []string{"users"})
	if err != nil {
		t.Fatalf("NewTokenIssuer() error = %v", err)
	}

	token, err := issuer.GenerateAccessToken("123", map[string]any{"role": "admin"})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := issuer.ParseAndValidate(token)
	if err != nil {
		t.Fatalf("ParseAndValidate() error = %v", err)
	}

	if claims["sub"] != "123" {
		t.Fatalf("sub claim = %v, want 123", claims["sub"])
	}
	if claims["role"] != "admin" {
		t.Fatalf("role claim = %v, want admin", claims["role"])
	}

	if ttl := issuer.AccessTokenTTL(); ttl <= 0 {
		t.Fatalf("AccessTokenTTL() = %v, want > 0", ttl)
	}
	if ttl := issuer.RefreshTokenTTL(); ttl <= 0 {
		t.Fatalf("RefreshTokenTTL() = %v, want > 0", ttl)
	}
}

func TestTokenIssuer_ParseAndValidateRejectsDifferentMethod(t *testing.T) {
	priv, pub := generateKeyPair(t)
	issuer, err := NewTokenIssuer(priv, pub, "svc", []string{"users"})
	if err != nil {
		t.Fatalf("NewTokenIssuer() error = %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "123",
		"iss": "svc",
		"aud": []string{"users"},
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := issuer.ParseAndValidate(tokenString); err == nil {
		t.Fatal("ParseAndValidate() expected error for invalid method")
	}
}

func TestLoadIssuerFromFiles(t *testing.T) {
	priv, pub := generateKeyPair(t)

	dir := t.TempDir()
	privPath := filepath.Join(dir, "key")
	pubPath := filepath.Join(dir, "key.pub")

	if err := os.WriteFile(privPath, priv, 0o600); err != nil {
		t.Fatalf("WriteFile() priv error = %v", err)
	}
	if err := os.WriteFile(pubPath, pub, 0o600); err != nil {
		t.Fatalf("WriteFile() pub error = %v", err)
	}

	issuer, err := LoadIssuerFromFiles(privPath, pubPath, "svc", []string{"users"})
	if err != nil {
		t.Fatalf("LoadIssuerFromFiles() error = %v", err)
	}
	if issuer == nil {
		t.Fatal("LoadIssuerFromFiles() returned nil issuer")
	}
}

func TestNewTokenIssuerInvalidKeys(t *testing.T) {
	_, err := NewTokenIssuer([]byte("not-a-key"), []byte("still-not"), "svc", nil)
	if err == nil {
		t.Fatal("expected error for invalid private key")
	}

	priv, pub := generateKeyPair(t)
	// Corrupt public key
	pub = pub[:len(pub)/2]
	if _, err := NewTokenIssuer(priv, pub, "svc", nil); err == nil {
		t.Fatal("expected error for invalid public key")
	}
}

func TestLoadIssuerFromFilesErrors(t *testing.T) {
	dir := t.TempDir()
	privPath := filepath.Join(dir, "priv")
	// Missing file should trigger error
	if _, err := LoadIssuerFromFiles(privPath, filepath.Join(dir, "pub"), "svc", nil); err == nil {
		t.Fatal("expected error when files are missing")
	}
}

func TestLoadIssuerFromFilesPublicKeyError(t *testing.T) {
	priv, _ := generateKeyPair(t)
	dir := t.TempDir()
	privPath := filepath.Join(dir, "priv.pem")
	if err := os.WriteFile(privPath, priv, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadIssuerFromFiles(privPath, filepath.Join(dir, "missing.pub"), "svc", nil); err == nil {
		t.Fatal("expected error for missing public key")
	}
}
