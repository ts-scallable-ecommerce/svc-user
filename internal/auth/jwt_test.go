package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	keyOnce    sync.Once
	cachedPriv []byte
	cachedPub  []byte
)

func getTestKeys(t *testing.T) ([]byte, []byte) {
	t.Helper()
	keyOnce.Do(func() {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatalf("generate key: %v", err)
		}
		privBytes := x509.MarshalPKCS1PrivateKey(key)
		cachedPriv = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
		pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			t.Fatalf("marshal public key: %v", err)
		}
		cachedPub = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	})
	return cachedPriv, cachedPub
}

func TestNewTokenIssuerAndGenerate(t *testing.T) {
	privPEM, pubPEM := getTestKeys(t)
	issuer, err := NewTokenIssuer(privPEM, pubPEM, "svc-user", []string{"api"})
	if err != nil {
		t.Fatalf("NewTokenIssuer returned error: %v", err)
	}

	token, err := issuer.GenerateAccessToken("user-123", map[string]any{"role": "member"})
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}
	claims, err := issuer.ParseAndValidate(token)
	if err != nil {
		t.Fatalf("ParseAndValidate error: %v", err)
	}
	if claims["sub"] != "user-123" {
		t.Fatalf("unexpected subject: %v", claims["sub"])
	}
	if claims["role"] != "member" {
		t.Fatalf("missing custom claim")
	}
	if ttl := issuer.RefreshTokenTTL(); ttl != 7*24*time.Hour {
		t.Fatalf("unexpected refresh ttl: %v", ttl)
	}
	if ttl := issuer.AccessTokenTTL(); ttl != 15*time.Minute {
		t.Fatalf("unexpected access ttl: %v", ttl)
	}
}

func TestParseAndValidateRejectsInvalidToken(t *testing.T) {
	priv, pub := getTestKeys(t)
	issuer, err := NewTokenIssuer(priv, pub, "svc-user", []string{"api"})
	if err != nil {
		t.Fatalf("NewTokenIssuer returned error: %v", err)
	}

	if _, err := issuer.ParseAndValidate("not-a-jwt"); err == nil {
		t.Fatalf("expected error for malformed token")
	}
}

func TestParseAndValidateUnexpectedMethod(t *testing.T) {
	priv, pub := getTestKeys(t)
	issuer, err := NewTokenIssuer(priv, pub, "svc-user", []string{"api"})
	if err != nil {
		t.Fatalf("NewTokenIssuer returned error: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "user"})
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString error: %v", err)
	}
	if _, err := issuer.ParseAndValidate(signed); err == nil {
		t.Fatalf("expected unexpected signing method error")
	}
}

func TestLoadIssuerFromFiles(t *testing.T) {
	priv, pub := getTestKeys(t)
	dir := t.TempDir()
	privPath := filepath.Join(dir, "private.pem")
	pubPath := filepath.Join(dir, "public.pem")
	if err := os.WriteFile(privPath, priv, 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
	if err := os.WriteFile(pubPath, pub, 0o600); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	issuer, err := LoadIssuerFromFiles(privPath, pubPath, "svc-user", []string{"api"})
	if err != nil {
		t.Fatalf("LoadIssuerFromFiles error: %v", err)
	}

	token, err := issuer.GenerateAccessToken("user-1", nil)
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}
	if _, err := issuer.ParseAndValidate(token); err != nil {
		t.Fatalf("ParseAndValidate error: %v", err)
	}
}

func TestLoadIssuerFromFilesMissing(t *testing.T) {
	if _, err := LoadIssuerFromFiles("missing", "missing", "svc-user", nil); err == nil {
		t.Fatalf("expected error when files missing")
	}
}

func TestNewTokenIssuerInvalidKeyMaterial(t *testing.T) {
	priv, pub := getTestKeys(t)
	if _, err := NewTokenIssuer([]byte("not a key"), pub, "svc-user", nil); err == nil {
		t.Fatalf("expected error for invalid private key")
	}
	if _, err := NewTokenIssuer(priv, []byte("not a key"), "svc-user", nil); err == nil {
		t.Fatalf("expected error for invalid public key")
	}
}

func TestGenerateAccessTokenMissingKey(t *testing.T) {
	issuer := &TokenIssuer{}
	if _, err := issuer.GenerateAccessToken("user", nil); err == nil {
		t.Fatalf("expected error when signing key missing")
	}
}
