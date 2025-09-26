package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
)

// Service orchestrates user business logic.
type Service struct {
	repo        Repository
	issuer      *auth.TokenIssuer
	roleStore   RoleStore
	revocations auth.TokenBlacklist
}

// NewService constructs the service dependencies.
func NewService(repo Repository, issuer *auth.TokenIssuer, roles RoleStore, revocations auth.TokenBlacklist) *Service {
	return &Service{repo: repo, issuer: issuer, roleStore: roles, revocations: revocations}
}

// RoleStore exposes RBAC operations required by the service.
type RoleStore interface {
	AssignRole(ctx context.Context, userID, role string) error
	ListRoles(ctx context.Context, userID string) ([]string, error)
	ResolvePermissions(ctx context.Context, userID string) ([]string, error)
	HasPermission(ctx context.Context, userID, permission string) (bool, error)
}

// RegisterRequest captures the inbound payload for creating a user.
type RegisterRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// AuthenticateRequest captures login information.
type AuthenticateRequest struct {
	Email    string
	Password string
}

// UpdateProfileRequest captures mutable profile fields.
type UpdateProfileRequest struct {
	FirstName string
	LastName  string
}

// RegisterResult describes the registration outcome.
type RegisterResult struct {
	UserID string
	Tokens TokenPair
}

// AuthenticateResult contains the authenticated user profile and tokens.
type AuthenticateResult struct {
	UserID string
	Tokens TokenPair
}

// TokenPair represents issued access and refresh tokens.
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresIn  int64
	RefreshExpiresIn int64
}

// Profile describes the public user profile.
type Profile struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Status    string
	Roles     []string
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
	ErrTokenInvalid       = errors.New("invalid token")
)

// Register orchestrates the basic user registration flow.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	email := strings.TrimSpace(req.Email)
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        strings.ToLower(email),
		PasswordHash: hash,
		FirstName:    sqlString(req.FirstName),
		LastName:     sqlString(req.LastName),
		Status:       "pending",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	tokens, err := s.issueTokens(user)
	if err != nil {
		return nil, err
	}

	if s.roleStore != nil {
		_ = s.roleStore.AssignRole(ctx, user.ID, "customer")
	}

	return &RegisterResult{UserID: user.ID, Tokens: *tokens}, nil
}

// Authenticate validates the email/password credentials and issues tokens.
func (s *Service) Authenticate(ctx context.Context, req AuthenticateRequest) (*AuthenticateResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	match, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
		return nil, ErrInvalidCredentials
	}

	if user.Status == "disabled" {
		return nil, ErrUserDisabled
	}

	tokens, err := s.issueTokens(user)
	if err != nil {
		return nil, err
	}

	return &AuthenticateResult{UserID: user.ID, Tokens: *tokens}, nil
}

// GetProfile returns the profile for a user ID.
func (s *Service) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles, err := s.listRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &Profile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
		Status:    user.Status,
		Roles:     roles,
	}, nil
}

// UpdateProfile updates mutable profile fields.
func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*Profile, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.FirstName = sqlString(req.FirstName)
	user.LastName = sqlString(req.LastName)

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return s.GetProfile(ctx, userID)
}

// ChangePassword verifies the current password and updates the stored hash.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	match, err := auth.VerifyPassword(currentPassword, user.PasswordHash)
	if err != nil || !match {
		return ErrInvalidCredentials
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	return s.repo.Update(ctx, user)
}

// AssignRole assigns a role to a user.
func (s *Service) AssignRole(ctx context.Context, userID, role string) error {
	if s.roleStore == nil {
		return errors.New("role store not configured")
	}
	return s.roleStore.AssignRole(ctx, userID, role)
}

// Permissions resolves permissions for a user.
func (s *Service) Permissions(ctx context.Context, userID string) ([]string, error) {
	if s.roleStore == nil {
		return nil, errors.New("role store not configured")
	}
	return s.roleStore.ResolvePermissions(ctx, userID)
}

// HasPermission checks if the user has the specified permission.
func (s *Service) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	if s.roleStore == nil {
		return false, errors.New("role store not configured")
	}
	return s.roleStore.HasPermission(ctx, userID, permission)
}

// Logout blacklists the supplied token until it would naturally expire.
func (s *Service) Logout(ctx context.Context, token string) error {
	if s.revocations == nil {
		return errors.New("token blacklist not configured")
	}

	claims, err := s.issuer.ParseAndValidate(token)
	if err != nil {
		return ErrTokenInvalid
	}

	expValue, ok := claims["exp"]
	if !ok {
		return ErrTokenInvalid
	}

	exp, err := parseExpiration(expValue)
	if err != nil {
		return ErrTokenInvalid
	}

	ttl := time.Until(time.Unix(exp, 0))
	if ttl <= 0 {
		ttl = time.Second
	}

	return s.revocations.Revoke(ctx, token, ttl)
}

// ParseSubject extracts the user ID from a JWT.
func (s *Service) ParseSubject(token string) (string, error) {
	return s.issuer.SubjectFromToken(token)
}

func (s *Service) issueTokens(user *User) (*TokenPair, error) {
	access, err := s.issuer.GenerateAccessToken(user.ID, map[string]any{
		"email": user.Email,
	})
	if err != nil {
		return nil, err
	}

	refresh, err := s.issuer.GenerateRefreshToken(user.ID, map[string]any{})
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      access,
		RefreshToken:     refresh,
		AccessExpiresIn:  int64(s.issuer.AccessTokenTTL().Seconds()),
		RefreshExpiresIn: int64(s.issuer.RefreshTokenTTL().Seconds()),
	}, nil
}

func (s *Service) listRoles(ctx context.Context, userID string) ([]string, error) {
	if s.roleStore == nil {
		return nil, nil
	}
	return s.roleStore.ListRoles(ctx, userID)
}

func sqlString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func validateEmail(email string) error {
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

func parseExpiration(value any) (int64, error) {
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, err
		}
		return i, nil
	default:
		return 0, fmt.Errorf("unsupported expiration type %T", value)
	}
}
