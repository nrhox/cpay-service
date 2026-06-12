package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func GenerateDynamicToken(secretKey string) string {
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(timestampStr))
	tokenHMAC := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	tokenTimestamp := base64.RawURLEncoding.EncodeToString([]byte(timestampStr))

	return fmt.Sprintf("%s.%s", tokenHMAC, tokenTimestamp)
}

func ValidateDynamicToken(combinedToken string, secretKey string) bool {
	parts := strings.Split(combinedToken, ".")
	if len(parts) != 2 {
		return false
	}
	clientTokenHMAC := parts[0]
	clientTimestampBase64 := parts[1]

	decodedTimestampBytes, err := base64.RawURLEncoding.DecodeString(clientTimestampBase64)
	if err != nil {
		return false
	}
	timestampStr := string(decodedTimestampBytes)

	clientTimestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}

	currentTime := time.Now().Unix()
	const timeTolerance int64 = 60
	if math.Abs(float64(currentTime-clientTimestamp)) > float64(timeTolerance) {
		return false // Token kedaluwarsa
	}

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(timestampStr))
	expectedTokenHMAC := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(clientTokenHMAC), []byte(expectedTokenHMAC)) == 1
}
