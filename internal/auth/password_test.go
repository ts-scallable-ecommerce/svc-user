package auth

import (
	"fmt"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("super-secret")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == "" {
		t.Fatalf("expected hash to be non-empty")
	}

	ok, err := VerifyPassword("super-secret", hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected password verification to succeed")
	}

	ok, err = VerifyPassword("wrong", hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}
	if ok {
		t.Fatalf("expected password verification to fail for incorrect password")
	}
}

func TestVerifyPasswordInvalidFormat(t *testing.T) {
	if _, err := VerifyPassword("anything", "invalid"); err == nil {
		t.Fatalf("expected error for invalid hash format")
	}
}

func TestVerifyPasswordUnsupportedVersion(t *testing.T) {
	hash := "argon2id$v=18$m=1,t=1,p=1$AAAA$AAAA"
	if _, err := VerifyPassword("pw", hash); err == nil {
		t.Fatalf("expected unsupported version error")
	}
}

func TestVerifyPasswordUnexpectedParameter(t *testing.T) {
	hash := "argon2id$v=19$m=1,t=1,q=1$AAAA$AAAA"
	if _, err := VerifyPassword("pw", hash); err == nil {
		t.Fatalf("expected error for unexpected parameter")
	}
}

func TestVerifyPasswordDecodeErrors(t *testing.T) {
	hash := "argon2id$v=19$m=1,t=1,p=1$????$AAAA"
	if _, err := VerifyPassword("pw", hash); err == nil {
		t.Fatalf("expected salt decode error")
	}

	hash = "argon2id$v=19$m=1,t=1,p=1$AAAA$????"
	if _, err := VerifyPassword("pw", hash); err == nil {
		t.Fatalf("expected hash decode error")
	}
}

func TestVerifyPasswordParseUintError(t *testing.T) {
	hash := "argon2id$v=19$m=abc,t=1,p=1$AAAA$AAAA"
	if _, err := VerifyPassword("pw", hash); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestVerifyPasswordMissingParameter(t *testing.T) {
	hash := "argon2id$v=19$m=1,t=1,p=0$AAAA$AAAA"
	if _, err := VerifyPassword("pw", hash); err == nil {
		t.Fatalf("expected error when parameters are zero")
	}
}

func TestHashPasswordRandomFailure(t *testing.T) {
	restore := OverrideRandomRead(func(b []byte) (int, error) {
		return 0, fmt.Errorf("rng failure")
	})
	defer restore()

	if _, err := HashPassword("pw"); err == nil {
		t.Fatalf("expected error when random generator fails")
	}
}
