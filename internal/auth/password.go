package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var randomRead = rand.Read

const (
	saltLength  = 16
	hashLength  = 32
	iterations  = 3
	memory      = 64 * 1024
	parallelism = 2
)

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
	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" {
		return false, fmt.Errorf("invalid hash format")
	}

	var mem uint32
	var iter uint32
	var par uint8
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &mem, &iter, &par); err != nil {
		return false, fmt.Errorf("invalid hash format")
	}
	saltB64 := parts[3]
	hashB64 := parts[4]

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	derived := argon2.IDKey([]byte(password), salt, iter, mem, uint8(par), uint32(len(hash)))

	if subtle.ConstantTimeCompare(hash, derived) == 1 {
		return true, nil
	}
	return false, nil
}
