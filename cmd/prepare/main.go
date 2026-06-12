package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
)

func generateSaltKey(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-.~"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes), nil
}

func convertDERToPEMBase64URL(blockType string, derBytes []byte) (string, error) {
	var buf bytes.Buffer
	block := &pem.Block{
		Type:  blockType,
		Bytes: derBytes,
	}

	err := pem.Encode(&buf, block)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf.Bytes()), nil
}

func main() {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("failed generate key pair: %v", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		log.Fatalf("failed marshal private key: %v", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		log.Fatalf("failed marshal public key: %v", err)
	}

	pubBase64URL, err := convertDERToPEMBase64URL("PUBLIC KEY", pubBytes)
	if err != nil {
		log.Fatalf("failed to process public key: %v", err)
	}

	privBase64URL, err := convertDERToPEMBase64URL("PRIVATE KEY", privBytes)
	if err != nil {
		log.Fatalf("failed to process private key: %v", err)
	}

	saltKey, err := generateSaltKey(80)
	if err != nil {
		log.Fatalf("failed generate salt key: %v", err)
	}

	fmt.Printf("PUBLIC_KEY: %s\n", pubBase64URL)
	fmt.Printf("PRIVATE_KEY: %s\n", privBase64URL)
	fmt.Printf("SALT_KEY: %s\n", saltKey)
}
