package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory      uint32 = 64 * 1024
	argon2Iterations  uint32 = 3
	argon2Parallelism uint8  = 2
	argon2SaltLength  uint32 = 16
	argon2KeyLength   uint32 = 32
)

// HashPassword hashes the raw password using Argon2id and returns a versioned encoded string.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password must not be blank")
	}

	salt, err := randomBytes(argon2SaltLength)
	if err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Iterations,
		argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// VerifyPassword compares a raw password against an encoded Argon2id hash.
func VerifyPassword(password, encoded string) (bool, error) {
	if password == "" || encoded == "" {
		return false, nil
	}

	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid encoded hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parse hash version: %w", err)
	}
	if version != argon2.Version {
		return false, fmt.Errorf("unsupported argon2 version %d", version)
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("parse hash params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	calculated := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(decodedHash)))
	if len(calculated) != len(decodedHash) {
		return false, nil
	}

	return subtle.ConstantTimeCompare(calculated, decodedHash) == 1, nil
}

func randomBytes(length uint32) ([]byte, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}
