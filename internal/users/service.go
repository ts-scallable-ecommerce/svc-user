package users

import (
	"context"
	"database/sql"

	"github.com/tasiuskenways/scalable-ecommerce/svc-user/internal/auth"
)

// Service orchestrates user business logic.
type Service struct {
	repo   Repository
	issuer *auth.TokenIssuer
}

// NewService constructs the service dependencies.
func NewService(repo Repository, issuer *auth.TokenIssuer) *Service {
	return &Service{repo: repo, issuer: issuer}
}

// RegisterRequest captures the inbound payload for creating a user.
type RegisterRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// RegisterResult describes the registration outcome.
type RegisterResult struct {
	UserID      string
	AccessToken string
	RefreshTTL  int64
}

// Register orchestrates the basic user registration flow.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Email:        req.Email,
		PasswordHash: hash,
		FirstName:    sqlString(req.FirstName),
		LastName:     sqlString(req.LastName),
		Status:       "pending",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.issuer.GenerateAccessToken(user.ID, map[string]any{
		"email": user.Email,
	})
	if err != nil {
		return nil, err
	}

	return &RegisterResult{
		UserID:      user.ID,
		AccessToken: token,
		RefreshTTL:  int64(s.issuer.RefreshTokenTTL().Seconds()),
	}, nil
}

func sqlString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
