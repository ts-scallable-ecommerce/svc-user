package users

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"sync"
	"testing"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
)

type stubRepository struct {
	created []*User
	err     error
}

func (s *stubRepository) Create(_ context.Context, u *User) error {
	if s.err != nil {
		return s.err
	}
	u.ID = "user-id"
	s.created = append(s.created, u)
	return nil
}

func (s *stubRepository) FindByEmail(context.Context, string) (*User, error) {
	return nil, errors.New("not implemented")
}
func (s *stubRepository) FindByID(context.Context, string) (*User, error) {
	return nil, errors.New("not implemented")
}
func (s *stubRepository) Update(context.Context, *User) error { return errors.New("not implemented") }

func TestServiceRegisterSuccess(t *testing.T) {
	priv, pub := testKeys(t)
	issuer, err := auth.NewTokenIssuer(priv, pub, "svc-user", []string{"api"})
	if err != nil {
		t.Fatalf("issuer error: %v", err)
	}

	repo := &stubRepository{}
	svc := NewService(repo, issuer)

	res, err := svc.Register(context.Background(), RegisterRequest{
		Email:     "user@example.com",
		Password:  "secret",
		FirstName: "Jane",
		LastName:  "Doe",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if res.UserID != "user-id" {
		t.Fatalf("expected propagated user id")
	}
	if res.AccessToken == "" {
		t.Fatalf("expected non-empty access token")
	}
	if res.RefreshTTL <= 0 {
		t.Fatalf("expected positive refresh ttl")
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected exactly one user created")
	}
	created := repo.created[0]
	if got := created.FirstName.String; got != "Jane" {
		t.Fatalf("unexpected first name: %s", got)
	}
	if got := created.LastName.String; got != "Doe" {
		t.Fatalf("unexpected last name: %s", got)
	}
	if created.Status != "pending" {
		t.Fatalf("expected pending status")
	}
	if created.PasswordHash == "" {
		t.Fatalf("expected password hash to be set")
	}
}

func TestServiceRegisterRepositoryFailure(t *testing.T) {
	priv, pub := testKeys(t)
	issuer, err := auth.NewTokenIssuer(priv, pub, "svc-user", []string{"api"})
	if err != nil {
		t.Fatalf("issuer error: %v", err)
	}

	repo := &stubRepository{err: errors.New("db failure")}
	svc := NewService(repo, issuer)

	if _, err := svc.Register(context.Background(), RegisterRequest{Email: "user@example.com", Password: "secret"}); err == nil {
		t.Fatalf("expected error from repository")
	}
	if len(repo.created) != 0 {
		t.Fatalf("unexpected user creation on failure")
	}
}

func TestServiceRegisterTokenFailure(t *testing.T) {
	repo := &stubRepository{}
	svc := NewService(repo, &auth.TokenIssuer{})

	if _, err := svc.Register(context.Background(), RegisterRequest{Email: "user@example.com", Password: "secret"}); err == nil {
		t.Fatalf("expected token generation failure")
	}
}

func TestSQLString(t *testing.T) {
	if v := sqlString(""); v.Valid {
		t.Fatalf("expected invalid NullString for empty input")
	}
	if v := sqlString("value"); !v.Valid || v.String != "value" {
		t.Fatalf("expected valid NullString for non-empty input")
	}
}

var (
	once       sync.Once
	privatePEM []byte
	publicPEM  []byte
)

func testKeys(t *testing.T) ([]byte, []byte) {
	t.Helper()
	once.Do(func() {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatalf("generate key: %v", err)
		}
		privBytes := x509.MarshalPKCS1PrivateKey(key)
		privatePEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
		pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			t.Fatalf("marshal public key: %v", err)
		}
		publicPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	})
	return privatePEM, publicPEM
}
