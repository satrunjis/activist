package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

const tokenBytes = 32

// NewOpaqueToken creates a cryptographically secure random token suitable for cookies.
func NewOpaqueToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate opaque token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashToken creates a deterministic SHA-256 hash for persistent session lookup.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// SessionExpired reports whether the session is unusable due to idle or absolute TTL.
func SessionExpired(now, idleExpiresAt, absoluteExpiresAt time.Time) bool {
	now = now.UTC()
	return !now.Before(idleExpiresAt.UTC()) || !now.Before(absoluteExpiresAt.UTC())
}

// NextIdleExpiry computes the next idle expiry while keeping it bounded by absolute expiry.
func NextIdleExpiry(now time.Time, idleTTL time.Duration, absoluteExpiresAt time.Time) time.Time {
	candidate := now.UTC().Add(idleTTL)
	if candidate.After(absoluteExpiresAt.UTC()) {
		return absoluteExpiresAt.UTC()
	}
	return candidate
}
