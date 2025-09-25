package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength  = 16
	hashLength  = 32
	iterations  = 3
	memory      = 64 * 1024
	parallelism = 2
)

var randomRead = rand.Read

// HashPassword derives an Argon2id hash for the supplied password.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := randomRead(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, hashLength)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memory, iterations, parallelism, encodedSalt, encodedHash), nil
}

// VerifyPassword compares a password with the encoded hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 || parts[0] != "argon2id" {
		return false, fmt.Errorf("invalid hash format")
	}
	if parts[1] != "v=19" {
		return false, fmt.Errorf("unsupported version")
	}

	params := strings.Split(parts[2], ",")
	var mem uint64
	var iter uint64
	var par uint64
	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			return false, fmt.Errorf("invalid hash format")
		}
		value, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return false, fmt.Errorf("invalid hash parameter: %w", err)
		}
		switch kv[0] {
		case "m":
			mem = value
		case "t":
			iter = value
		case "p":
			par = value
		default:
			return false, fmt.Errorf("unexpected hash parameter: %s", kv[0])
		}
	}
	if mem == 0 || iter == 0 || par == 0 {
		return false, fmt.Errorf("invalid hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	derived := argon2.IDKey([]byte(password), salt, uint32(iter), uint32(mem), uint8(par), uint32(len(hash)))

	if subtle.ConstantTimeCompare(hash, derived) == 1 {
		return true, nil
	}
	return false, nil
}

// OverrideRandomRead replaces the randomness source; primarily useful for tests.
func OverrideRandomRead(fn func([]byte) (int, error)) func() {
	original := randomRead
	randomRead = fn
	return func() { randomRead = original }
}
