package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const tokenBytes = 32

// NewOpaqueToken creates a URL-safe random bearer token.
func NewOpaqueToken() (string, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// TokenHash returns a fixed-size digest suitable for indexed database lookup.
// Raw bearer tokens must never be persisted or logged.
func TokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
