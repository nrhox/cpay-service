package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// GenerateRandomToken returns base64url raw token
func GenerateRandomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// returns hex/bytes or base64 of SHA256(token)
// store this hash in DB, never store raw token
func HashTokenForStorage(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
