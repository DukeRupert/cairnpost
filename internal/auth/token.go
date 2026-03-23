package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// GenerateSessionToken creates a random token and its SHA-256 hash.
// The raw token goes in the cookie; the hash goes in the database.
func GenerateSessionToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.URLEncoding.EncodeToString(b)
	hash = HashToken(raw)
	return raw, hash, nil
}

// HashToken returns the hex-encoded SHA-256 hash of a raw token.
func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
