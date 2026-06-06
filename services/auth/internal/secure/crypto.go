// Package secure contains small cryptographic helpers used by auth.
package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptString encrypts plaintext with AES-256-GCM and returns a base64 token.
func EncryptString(key []byte, plaintext string) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("secure: AES-256 key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("secure: create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("secure: create GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("secure: random nonce: %w", err)
	}
	out := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(out), nil
}

// DecryptString decrypts a base64 AES-256-GCM payload.
func DecryptString(key []byte, ciphertext string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("secure: decode ciphertext: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("secure: create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("secure: create GCM: %w", err)
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("secure: ciphertext too short")
	}
	nonce, payload := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", fmt.Errorf("secure: decrypt: %w", err)
	}
	return string(plain), nil
}

// SHA256Base64 returns a URL-safe SHA-256 digest for lookup keys.
func SHA256Base64(value string) string {
	sum := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
