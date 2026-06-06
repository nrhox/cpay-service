package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestGenerateAndValidateDynamicToken_Success(t *testing.T) {
	secretKey := "test-secret-key-123"
	token := GenerateDynamicToken(secretKey)

	if token == "" {
		t.Fatal("generated token is empty")
	}

	isValid := ValidateDynamicToken(token, secretKey)
	if !isValid {
		t.Error("valid token failed verification")
	}
}

func TestValidateDynamicToken_Cases(t *testing.T) {
	secretKey := "6h8rt4pjXiUU2nl9hHfR0JuIVdgbUn1lByadaBAasUMhf1-SS-8FYWhZTneQJZjj7w31vg~KVtcVJ92S"
	wrongSecret := "6h8rt4pjXiUU2nl9hHfddIVddgbUn1lByadaBAasUMhf1-SS-8FYWhZTneQJZjj7w31vg~KVtcVJ92S"

	validToken := GenerateDynamicToken(secretKey)

	expiredTimestamp := time.Now().Unix() - 120
	expiredTimeStr := strconv.FormatInt(expiredTimestamp, 10)
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(expiredTimeStr))
	expiredHMAC := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	expiredTokenTimestamp := base64.RawURLEncoding.EncodeToString([]byte(expiredTimeStr))
	expiredToken := fmt.Sprintf("%s.%s", expiredHMAC, expiredTokenTimestamp)

	tests := []struct {
		name         string
		token        string
		secret       string
		expectResult bool
	}{
		{
			name:         "Valid Token Correct Secret",
			token:        validToken,
			secret:       secretKey,
			expectResult: true,
		},
		{
			name:         "Incorrect Secret",
			token:        validToken,
			secret:       wrongSecret,
			expectResult: false,
		},
		{
			name:         "Expired Token",
			token:        expiredToken,
			secret:       secretKey,
			expectResult: false,
		},
		{
			name:         "Invalid Format No Dot",
			token:        "invalidtokenstring",
			secret:       secretKey,
			expectResult: false,
		},
		{
			name:         "Invalid Format Multiple Dots",
			token:        "part1.part2.part3",
			secret:       secretKey,
			expectResult: false,
		},
		{
			name:         "Timestamp Not Base64",
			token:        "hmacpart.invalid-base64-!",
			secret:       secretKey,
			expectResult: false,
		},
		{
			name:         "Timestamp Not Numeric",
			token:        "hmacpart." + base64.RawURLEncoding.EncodeToString([]byte("abc")),
			secret:       secretKey,
			expectResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateDynamicToken(tt.token, tt.secret)
			if result != tt.expectResult {
				t.Errorf("expected %v, got %v for case: %s", tt.expectResult, result, tt.name)
			}
		})
	}
}
