package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const defaultEncryptionKey = "employee-card-admin-secret-key-32b"

// getAESKey derives a 32-byte key from ENCRYPTION_KEY environment variable using SHA-256
func getAESKey() []byte {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		keyStr = defaultEncryptionKey
	}
	hash := sha256.Sum256([]byte(keyStr))
	return hash[:]
}

// Encrypt encrypts a plaintext string using AES-256-GCM and returns a base64 encoded string
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key := getAESKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal appends the ciphertext to nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64 encoded ciphertext string using AES-256-GCM
func Decrypt(encodedCiphertext string) (string, error) {
	if encodedCiphertext == "" {
		return "", nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}

	key := getAESKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}

	return string(plaintext), nil
}

// DecryptedIdentity represents the payload decrypted from an identity JWE token
type DecryptedIdentity struct {
	NID   string `json:"nid"`
	DOB   string `json:"dob"`
	Email string `json:"email"`
}

// DecryptJWEIdentity decrypts a compact JWE (RSA-OAEP-256 + A256GCM) using a PEM-encoded RSA private key
func DecryptJWEIdentity(token string, privateKeyPEM string) (*DecryptedIdentity, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 5 {
		return nil, errors.New("invalid JWE compact format: expected 5 parts")
	}

	headerB64 := parts[0]
	encKeyB64 := parts[1]
	ivB64 := parts[2]
	cipherB64 := parts[3]
	tagB64 := parts[4]

	// 1. Parse PEM private key (supports PKCS#8 and PKCS#1)
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}

	var rsaPrivKey *rsa.PrivateKey
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		var ok bool
		rsaPrivKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("parsed key is not an RSA private key")
		}
	} else if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		rsaPrivKey = key
	} else {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	// 2. Decode Encrypted Key
	encKeyBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(encKeyB64, "="))
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	// 3. Decrypt CEK with RSA-OAEP-256
	cek, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaPrivKey, encKeyBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt CEK via RSA-OAEP: %w", err)
	}

	// 4. Decode IV, Ciphertext, and Tag
	ivBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(ivB64, "="))
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	cipherBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(cipherB64, "="))
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	tagBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(tagB64, "="))
	if err != nil {
		return nil, fmt.Errorf("failed to decode tag: %w", err)
	}

	// 5. Decrypt with AES-GCM
	blockCipher, err := aes.NewCipher(cek)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	ciphertextWithTag := append(cipherBytes, tagBytes...)
	aad := []byte(headerB64)

	plaintext, err := gcm.Open(nil, ivBytes, ciphertextWithTag, aad)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES-GCM ciphertext: %w", err)
	}

	var identity DecryptedIdentity
	if err := json.Unmarshal(plaintext, &identity); err != nil {
		identity.Email = strings.Trim(strings.TrimSpace(string(plaintext)), `"'`)
	}

	return &identity, nil
}
