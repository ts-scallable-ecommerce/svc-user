package users_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/users"
)

func TestRegisterAuthenticateAndProfileFlow(t *testing.T) {
	svc, repo, roles := newTestService(t)
	ctx := context.Background()

	res, err := svc.Register(ctx, users.RegisterRequest{
		Email:     "test@example.com",
		Password:  "Password!2",
		FirstName: "Jane",
		LastName:  "Doe",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if res.UserID == "" || res.Tokens.AccessToken == "" {
		t.Fatal("expected tokens and user id")
	}

	if _, ok := repo.byID[res.UserID]; !ok {
		t.Fatalf("user not persisted: %s", res.UserID)
	}
	if !roles.hasRole(res.UserID, "customer") {
		t.Fatalf("expected default customer role assigned")
	}

	authRes, err := svc.Authenticate(ctx, users.AuthenticateRequest{Email: "test@example.com", Password: "Password!2"})
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if authRes.Tokens.AccessToken == "" || authRes.Tokens.RefreshToken == "" {
		t.Fatal("expected issued tokens")
	}

	profile, err := svc.GetProfile(ctx, res.UserID)
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if profile.Email != "test@example.com" || profile.FirstName != "Jane" {
		t.Fatalf("unexpected profile data: %+v", profile)
	}

	updated, err := svc.UpdateProfile(ctx, res.UserID, users.UpdateProfileRequest{FirstName: "Janet", LastName: "Smith"})
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.FirstName != "Janet" || updated.LastName != "Smith" {
		t.Fatalf("profile not updated: %+v", updated)
	}

	if err := svc.ChangePassword(ctx, res.UserID, "Password!2", "NewPassword!3"); err != nil {
		t.Fatalf("change password: %v", err)
	}
	if _, err := svc.Authenticate(ctx, users.AuthenticateRequest{Email: "test@example.com", Password: "Password!2"}); err == nil {
		t.Fatal("expected old password to fail authentication")
	}

	roles.assign(res.UserID, "admin")
	roles.addPermission("admin", "roles:assign")
	allowed, err := svc.HasPermission(ctx, res.UserID, "roles:assign")
	if err != nil {
		t.Fatalf("has permission: %v", err)
	}
	if !allowed {
		t.Fatal("expected permission")
	}

	perms, err := svc.Permissions(ctx, res.UserID)
	if err != nil {
		t.Fatalf("permissions: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("expected permissions list")
	}
}

func TestRegisterValidation(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()

	if _, err := svc.Register(ctx, users.RegisterRequest{Email: "invalid", Password: "short"}); err == nil {
		t.Fatal("expected validation error for invalid email and password")
	}
}

func TestAuthenticateInvalidCredentials(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()

	if _, err := svc.Authenticate(ctx, users.AuthenticateRequest{Email: "missing@example.com", Password: "none"}); err == nil {
		t.Fatal("expected error for unknown user")
	}
}

func newTestService(t *testing.T) (*users.Service, *memoryRepo, *memoryRoles) {
	t.Helper()

	repo := newMemoryRepo()
	roles := newMemoryRoles()

	priv, pub := generateKeyPair(t)
	issuer, err := auth.NewTokenIssuer(priv, pub, "test", []string{"api"})
	if err != nil {
		t.Fatalf("new token issuer: %v", err)
	}

	svc := users.NewService(repo, issuer, roles)
	return svc, repo, roles
}

type memoryRepo struct {
	mu      sync.RWMutex
	byID    map[string]*users.User
	byEmail map[string]*users.User
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{byID: make(map[string]*users.User), byEmail: make(map[string]*users.User)}
}

func (r *memoryRepo) Create(_ context.Context, u *users.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byEmail[u.Email]; exists {
		return users.ErrInvalidCredentials
	}
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	clone := *u
	r.byID[u.ID] = &clone
	r.byEmail[u.Email] = &clone
	return nil
}

func (r *memoryRepo) FindByEmail(_ context.Context, email string) (*users.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.byEmail[email]; ok {
		clone := *u
		return &clone, nil
	}
	return nil, users.ErrNotFound
}

func (r *memoryRepo) FindByID(_ context.Context, id string) (*users.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.byID[id]; ok {
		clone := *u
		return &clone, nil
	}
	return nil, users.ErrNotFound
}

func (r *memoryRepo) Update(_ context.Context, u *users.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[u.ID]; !ok {
		return users.ErrNotFound
	}
	clone := *u
	r.byID[u.ID] = &clone
	r.byEmail[u.Email] = &clone
	return nil
}

type memoryRoles struct {
	mu          sync.RWMutex
	roles       map[string]map[string]struct{}
	permissions map[string]map[string]struct{}
}

func newMemoryRoles() *memoryRoles {
	return &memoryRoles{
		roles:       make(map[string]map[string]struct{}),
		permissions: make(map[string]map[string]struct{}),
	}
}

func (m *memoryRoles) AssignRole(_ context.Context, userID, role string) error {
	m.assign(userID, role)
	return nil
}

func (m *memoryRoles) assign(userID, role string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.roles[userID]; !ok {
		m.roles[userID] = make(map[string]struct{})
	}
	m.roles[userID][role] = struct{}{}
}

func (m *memoryRoles) ListRoles(_ context.Context, userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	roles := m.roles[userID]
	out := make([]string, 0, len(roles))
	for role := range roles {
		out = append(out, role)
	}
	return out, nil
}

func (m *memoryRoles) ResolvePermissions(_ context.Context, userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	roles := m.roles[userID]
	var perms []string
	for role := range roles {
		for perm := range m.permissions[role] {
			perms = append(perms, perm)
		}
	}
	return perms, nil
}

func (m *memoryRoles) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	perms, _ := m.ResolvePermissions(ctx, userID)
	for _, perm := range perms {
		if perm == permission {
			return true, nil
		}
	}
	return false, nil
}

func (m *memoryRoles) addPermission(role, permission string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.permissions[role]; !ok {
		m.permissions[role] = make(map[string]struct{})
	}
	m.permissions[role][permission] = struct{}{}
}

func (m *memoryRoles) hasRole(userID, role string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	roles := m.roles[userID]
	_, ok := roles[role]
	return ok
}

func generateKeyPair(t *testing.T) ([]byte, []byte) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	return privPEM, pubPEM
}
