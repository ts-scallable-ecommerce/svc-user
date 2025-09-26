package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenIssuer issues and validates JWT access and refresh tokens.
type TokenIssuer struct {
	signingKey *rsa.PrivateKey
	verifyKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
	audience   []string
}

// NewTokenIssuer constructs an issuer from PEM encoded key pairs.
func NewTokenIssuer(privateKeyPEM, publicKeyPEM []byte, issuer string, audience []string) (*TokenIssuer, error) {
	priv, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	pub, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &TokenIssuer{
		signingKey: priv,
		verifyKey:  pub,
		accessTTL:  15 * time.Minute,
		refreshTTL: 7 * 24 * time.Hour,
		issuer:     issuer,
		audience:   audience,
	}, nil
}

// LoadIssuerFromFiles reads PEM files from disk and constructs an issuer.
func LoadIssuerFromFiles(privatePath, publicPath, issuer string, audience []string) (*TokenIssuer, error) {
	priv, err := os.ReadFile(privatePath)
	if err != nil {
		return nil, err
	}
	pub, err := os.ReadFile(publicPath)
	if err != nil {
		return nil, err
	}
	return NewTokenIssuer(priv, pub, issuer, audience)
}

// GenerateAccessToken issues a signed JWT for the supplied claims.
func (t *TokenIssuer) GenerateAccessToken(subject string, claims map[string]any) (string, error) {
	return t.generateToken(subject, claims, t.accessTTL, "access")
}

// GenerateRefreshToken issues a signed JWT refresh token for the supplied subject.
func (t *TokenIssuer) GenerateRefreshToken(subject string, claims map[string]any) (string, error) {
	return t.generateToken(subject, claims, t.refreshTTL, "refresh")
}

func (t *TokenIssuer) generateToken(subject string, claims map[string]any, ttl time.Duration, tokenType string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": subject,
		"iss": t.issuer,
		"aud": t.audience,
		"exp": now.Add(ttl).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"typ": tokenType,
	})
	for k, v := range claims {
		token.Claims.(jwt.MapClaims)[k] = v
	}
	return token.SignedString(t.signingKey)
}

// ParseAndValidate validates an incoming token string and returns the claims.
func (t *TokenIssuer) ParseAndValidate(tokenString string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return t.verifyKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// AccessTokenTTL exposes the configured TTL for access tokens.
func (t *TokenIssuer) AccessTokenTTL() time.Duration {
	return t.accessTTL
}

// RefreshTokenTTL exposes the configured TTL for refresh tokens.
func (t *TokenIssuer) RefreshTokenTTL() time.Duration {
	return t.refreshTTL
}

// SubjectFromToken extracts the authenticated subject identifier from a token string.
func (t *TokenIssuer) SubjectFromToken(tokenString string) (string, error) {
	claims, err := t.ParseAndValidate(tokenString)
	if err != nil {
		return "", err
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", fmt.Errorf("token missing subject")
	}
	return sub, nil
}
