package users

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
)

type stubRepository struct {
	createFn func(ctx context.Context, u *User) error
}

func (s stubRepository) Create(ctx context.Context, u *User) error {
	if s.createFn != nil {
		return s.createFn(ctx, u)
	}
	u.ID = "generated-id"
	return nil
}

func (stubRepository) FindByEmail(ctx context.Context, email string) (*User, error) { return nil, nil }
func (stubRepository) FindByID(ctx context.Context, id string) (*User, error)       { return nil, nil }
func (stubRepository) Update(ctx context.Context, u *User) error                    { return nil }

func createIssuer(t *testing.T) *auth.TokenIssuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	priv := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pub := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)})

	issuer, err := auth.NewTokenIssuer(priv, pub, "svc-user", []string{"api"})
	if err != nil {
		t.Fatalf("NewTokenIssuer error: %v", err)
	}
	return issuer
}

func TestServiceRegisterSuccess(t *testing.T) {
	repo := stubRepository{}
	issuer := createIssuer(t)
	svc := NewService(repo, issuer)

	res, err := svc.Register(context.Background(), RegisterRequest{Email: "user@example.com", Password: "password"})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if res.UserID != "generated-id" {
		t.Fatalf("expected user ID from repository")
	}
	if res.AccessToken == "" {
		t.Fatalf("expected access token to be returned")
	}
	if res.RefreshTTL <= 0 {
		t.Fatalf("expected refresh TTL to be positive")
	}
}

func TestServiceRegisterRepoError(t *testing.T) {
	repo := stubRepository{createFn: func(ctx context.Context, u *User) error {
		return errors.New("db down")
	}}
	issuer := createIssuer(t)
	svc := NewService(repo, issuer)

	if _, err := svc.Register(context.Background(), RegisterRequest{Email: "user@example.com", Password: "password"}); err == nil {
		t.Fatalf("expected repository error to be returned")
	}
}

func TestServiceRegisterHashFailure(t *testing.T) {
	repo := stubRepository{}
	issuer := createIssuer(t)
	svc := NewService(repo, issuer)

	restore := auth.OverrideRandomRead(func(b []byte) (int, error) {
		return 0, errors.New("rng failure")
	})
	defer restore()

	if _, err := svc.Register(context.Background(), RegisterRequest{Email: "user@example.com", Password: "password"}); err == nil {
		t.Fatalf("expected hashing error to propagate")
	}
}

func TestSQLStringHelper(t *testing.T) {
	if out := sqlString(""); out.Valid {
		t.Fatalf("expected empty string to produce invalid sql.NullString")
	}
	if out := sqlString("value"); !out.Valid || out.String != "value" {
		t.Fatalf("expected value to be preserved")
	}
}
