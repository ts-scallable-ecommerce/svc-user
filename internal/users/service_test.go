package users

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
)

type stubRepository struct {
	created []*User
	err     error
}

func (s *stubRepository) Create(ctx context.Context, u *User) error {
	if s.err != nil {
		return s.err
	}
	s.created = append(s.created, u)
	if u.ID == "" {
		u.ID = "generated-id"
	}
	return nil
}

func (s *stubRepository) FindByEmail(context.Context, string) (*User, error) {
	return nil, nil
}

func (s *stubRepository) FindByID(context.Context, string) (*User, error) {
	return nil, nil
}

func (s *stubRepository) Update(context.Context, *User) error {
	return nil
}

func TestServiceRegister(t *testing.T) {
	repo := &stubRepository{}
	issuer := newTokenIssuer(t)
	svc := NewService(repo, issuer)

	result, err := svc.Register(context.Background(), RegisterRequest{
		Email:     "user@example.com",
		Password:  "password",
		FirstName: "John",
		LastName:  "Doe",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if result.UserID == "" {
		t.Fatalf("Register() returned empty user id")
	}
	if result.AccessToken == "" {
		t.Fatalf("Register() returned empty access token")
	}
	if result.RefreshTTL != int64(issuer.RefreshTokenTTL().Seconds()) {
		t.Fatalf("RefreshTTL mismatch: %d vs %d", result.RefreshTTL, int64(issuer.RefreshTokenTTL().Seconds()))
	}

	if len(repo.created) != 1 {
		t.Fatalf("expected repository Create to be called once, got %d", len(repo.created))
	}
	if repo.created[0].PasswordHash == "password" {
		t.Fatalf("password hash should not equal plain password")
	}
	if repo.created[0].FirstName.Valid == false || repo.created[0].FirstName.String != "John" {
		t.Fatalf("expected FirstName to be populated")
	}
	if repo.created[0].LastName.Valid == false || repo.created[0].LastName.String != "Doe" {
		t.Fatalf("expected LastName to be populated")
	}
	if repo.created[0].Status != "pending" {
		t.Fatalf("expected status pending, got %s", repo.created[0].Status)
	}
}

func TestSQLStringHelper(t *testing.T) {
	if got := sqlString(""); got.Valid {
		t.Fatalf("sqlString empty should be invalid, got %+v", got)
	}
	if got := sqlString("value"); !got.Valid || got.String != "value" {
		t.Fatalf("sqlString populated mismatch: %+v", got)
	}
}

func TestServiceRegisterRepositoryError(t *testing.T) {
	repo := &stubRepository{err: fmt.Errorf("db down")}
	issuer := newTokenIssuer(t)
	svc := NewService(repo, issuer)

	if _, err := svc.Register(context.Background(), RegisterRequest{Email: "user@example.com", Password: "password"}); err == nil {
		t.Fatal("expected repository error to propagate")
	}
	if len(repo.created) != 0 {
		t.Fatalf("expected no users created when repo errors")
	}
}

func newTokenIssuer(t *testing.T) *auth.TokenIssuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	priv := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pub := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)})
	issuer, err := auth.NewTokenIssuer(priv, pub, "svc", []string{"users"})
	if err != nil {
		t.Fatalf("NewTokenIssuer() error = %v", err)
	}
	return issuer
}
