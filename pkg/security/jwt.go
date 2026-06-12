package security

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nrhox/cpay-service/internal/constants"
)

var (
	ErrWrongEd25519    = errors.New("is not key Ed25519")
	ErrFailedDecodePEM = errors.New("failed decode PEM")
)

type TokenManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewTokenManager(pemPrivateKey string, pemPublicKey string) *TokenManager {
	privateKey, err := parsePrivateKey([]byte(pemPrivateKey))
	if err != nil {
		panic(err.Error())
	}

	publicKey, err := parsePublicKey([]byte(pemPublicKey))
	if err != nil {
		panic(err.Error())
	}

	return &TokenManager{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

type AuthPayload struct {
	UserID string         `json:"user_id"`
	RoleId constants.Role `json:"role_id"`
}

type AuthClaim struct {
	jwt.RegisteredClaims
	AuthPayload
}

func (j *TokenManager) Sign(p AuthPayload, d time.Duration) (string, error) {
	claims := AuthClaim{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(d)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		AuthPayload: p,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(j.privateKey)
}

func (j *TokenManager) Verify(tokenString string) (*AuthPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaim{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}

		return j.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*AuthClaim); ok {
		return &claims.AuthPayload, nil
	}

	return nil, jwt.ErrTokenInvalidId
}

func parsePrivateKey(pemData []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePEM
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := priv.(ed25519.PrivateKey)
	if !ok {
		return nil, ErrWrongEd25519
	}

	return key, nil
}

func parsePublicKey(pemData []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, ErrFailedDecodePEM
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, ErrWrongEd25519
	}

	return key, nil
}
