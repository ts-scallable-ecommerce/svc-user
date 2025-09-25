package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hashed, err := HashPassword("p@ssw0rd")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if !strings.HasPrefix(hashed, "argon2id$") {
		t.Fatalf("unexpected hash format: %s", hashed)
	}
	ok, err := VerifyPassword("p@ssw0rd", hashed)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected verification success")
	}

	ok, err = VerifyPassword("other", hashed)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if ok {
		t.Fatalf("expected verification failure for incorrect password")
	}
}

func TestHashPasswordRandomReadError(t *testing.T) {
	original := randomRead
	defer func() { randomRead = original }()

	randomRead = func([]byte) (int, error) {
		return 0, errors.New("entropy error")
	}

	if _, err := HashPassword("secret"); err == nil {
		t.Fatalf("expected error when random reader fails")
	}
}

func TestVerifyPasswordInvalidFormats(t *testing.T) {
	if _, err := VerifyPassword("x", "invalid-format"); err == nil {
		t.Fatalf("expected error for invalid hash format")
	}

	malformed := "argon2id$v=19$m=65536,t=3,p=2$not-base64$***"
	if _, err := VerifyPassword("x", malformed); err == nil {
		t.Fatalf("expected error for malformed hash components")
	}

	badHash := "argon2id$v=19$m=65536,t=3,p=2$bm90LXNhbHQ$***"
	if _, err := VerifyPassword("x", badHash); err == nil {
		t.Fatalf("expected error for malformed hash data")
	}
}
