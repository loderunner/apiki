package keychain

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "apiki"
	accountName = "encryption-key"
)

// Keychain defines the interface for storing and retrieving encryption keys.
type Keychain interface {
	Store(key []byte) error
	Retrieve() ([]byte, error)
	Delete() error
}

type contextKey struct{}

// WithKeychain returns a context with the given keychain for injection.
func WithKeychain(ctx context.Context, kc Keychain) context.Context {
	return context.WithValue(ctx, contextKey{}, kc)
}

func fromContext(ctx context.Context) Keychain {
	if kc, ok := ctx.Value(contextKey{}).(Keychain); ok {
		return kc
	}
	return osKeychain{}
}

type osKeychain struct{}

func (osKeychain) Store(key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf(
			"invalid key size: expected 32 bytes, got %d",
			len(key),
		)
	}

	encoded := base64.StdEncoding.EncodeToString(key)
	if err := keyring.Set(serviceName, accountName, encoded); err != nil {
		return fmt.Errorf("failed to store key in keychain: %w", err)
	}

	return nil
}

func (osKeychain) Retrieve() ([]byte, error) {
	encoded, err := keyring.Get(serviceName, accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key from keychain: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf(
			"invalid key size: expected 32 bytes, got %d",
			len(key),
		)
	}

	return key, nil
}

func (osKeychain) Delete() error {
	err := keyring.Delete(serviceName, accountName)
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("failed to delete keychain item: %w", err)
	}

	return nil
}

// Store stores a 32-byte encryption key using the keychain from context.
// On macOS, this uses the macOS Keychain API.
// On Linux, this uses D-Bus Secret Service (GNOME Keyring/KWallet).
func Store(ctx context.Context, key []byte) error {
	return fromContext(ctx).Store(key)
}

// Retrieve retrieves the encryption key using the keychain from context.
// On macOS, this uses the macOS Keychain API.
// On Linux, this uses D-Bus Secret Service.
func Retrieve(ctx context.Context) ([]byte, error) {
	return fromContext(ctx).Retrieve()
}

// Delete removes the encryption key using the keychain from context.
func Delete(ctx context.Context) error {
	return fromContext(ctx).Delete()
}
