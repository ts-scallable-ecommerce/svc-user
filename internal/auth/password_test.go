package auth_test

import (
	"testing"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := auth.HashPassword("Sup3rSecret!")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if hash == "" {
		t.Fatal("expected hash to be non-empty")
	}

	ok, err := auth.VerifyPassword("Sup3rSecret!", hash)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}

	ok, err = auth.VerifyPassword("wrong", hash)
	if err != nil {
		t.Fatalf("verify password mismatch: %v", err)
	}
	if ok {
		t.Fatal("expected mismatch to fail")
	}
}
