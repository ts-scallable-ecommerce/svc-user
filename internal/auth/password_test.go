package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	encoded, err := HashPassword("super-secret")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	ok, err := VerifyPassword("super-secret", encoded)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if !ok {
		t.Fatalf("VerifyPassword() returned false, want true")
	}

	mismatch, err := VerifyPassword("wrong", encoded)
	if err != nil {
		t.Fatalf("VerifyPassword() mismatch error = %v", err)
	}
	if mismatch {
		t.Fatalf("VerifyPassword() mismatch = true, want false")
	}
}

func TestVerifyPasswordInvalidFormat(t *testing.T) {
	if _, err := VerifyPassword("pw", "invalid-format"); err == nil {
		t.Fatal("VerifyPassword() expected format error")
	}
}

func TestVerifyPasswordInvalidBase64(t *testing.T) {
	hash := "argon2id$v=19$m=65536,t=3,p=2$not-base64$also-bad"
	if _, err := VerifyPassword("pw", hash); err == nil {
		t.Fatal("VerifyPassword() expected decode error")
	}
}
