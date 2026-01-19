package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	// SaltSize is the size of the salt in bytes (16 bytes = 128 bits)
	SaltSize = 16

	// KeySize is the size of the encryption key in bytes (32 bytes = 256 bits)
	KeySize = 32

	// NonceSize is the size of the nonce for AES-GCM (12 bytes recommended)
	NonceSize = 12

	// TagSize is the size of the authentication tag for AES-GCM (16 bytes)
	TagSize = 16

	// Argon2id parameters
	argon2Memory      = 64 * 1024 // 64 MiB
	argon2Iterations  = 3
	argon2Parallelism = 4

	// Prefix for encrypted values
	prefix    = "enc:v1:"
	prefixLen = len(prefix)
)

// GenerateSalt generates a random salt for Argon2id key derivation.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	n, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	if n != SaltSize {
		return nil, fmt.Errorf(
			"failed to generate salt: expected %d bytes, got %d",
			SaltSize,
			n,
		)
	}
	return salt, nil
}

// DeriveKey derives a 32-byte encryption key from a password using Argon2id.
func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		argon2Iterations,
		argon2Memory,
		argon2Parallelism,
		KeySize,
	)
}

// ComputeVerifier computes an HMAC-SHA256 verifier for fast password
// validation.
// The verifier is HMAC-SHA256(key, salt).
func ComputeVerifier(key []byte, salt []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(salt)
	return mac.Sum(nil)
}

// VerifyPassword verifies a password against a verifier.
func VerifyPassword(password string, salt []byte, verifier []byte) bool {
	key := DeriveKey(password, salt)
	computed := ComputeVerifier(key, salt)
	return hmac.Equal(computed, verifier)
}

// GenerateKey generates a random 32-byte encryption key.
func GenerateKey() ([]byte, error) {
	key := make([]byte, KeySize)
	n, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	if n != KeySize {
		return nil, fmt.Errorf(
			"failed to generate key: expected %d bytes, got %d", KeySize, n,
		)
	}
	return key, nil
}

// Encrypt encrypts a plaintext value using AES-256-GCM.
// Returns a base64-encoded string with format:
// "enc:v1:base64(nonce||ciphertext||tag)"
func Encrypt(key []byte, plaintext string) (string, error) {
	if len(key) != KeySize {
		return "", errors.New("invalid key size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, NonceSize)
	n, err := rand.Read(nonce)
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	if n != NonceSize {
		return "", fmt.Errorf(
			"failed to generate nonce: expected %d bytes, got %d", NonceSize, n,
		)
	}

	ciphertext := aead.Seal(nonce, nonce, []byte(plaintext), nil)

	// Format: nonce (12 bytes) || ciphertext || tag (16 bytes)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return fmt.Sprintf("%s%s", prefix, encoded), nil
}

// IsEncrypted returns true if the value has the encryption prefix.
func IsEncrypted(value string) bool {
	return len(value) >= prefixLen && value[:prefixLen] == prefix
}

// Decrypt decrypts an encrypted value.
// Expects format: "enc:v1:base64(nonce||ciphertext||tag)"
func Decrypt(key []byte, encrypted string) (string, error) {
	if len(key) != KeySize {
		return "", errors.New("invalid key size")
	}

	// Check prefix
	if !IsEncrypted(encrypted) {
		return "", errors.New("invalid encryption format")
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encrypted[prefixLen:])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := data[:NonceSize]
	ciphertext := data[NonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
