package keychain

import (
	"encoding/base64"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "apiki"
	accountName = "encryption-key"
)

// Store stores a 32-byte encryption key in the OS keychain.
// On macOS, this uses the macOS Keychain API.
// On Linux, this uses D-Bus Secret Service (GNOME Keyring/KWallet).
func Store(key []byte) error {
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

// Retrieve retrieves the encryption key from the OS keychain.
// On macOS, this uses the macOS Keychain API.
// On Linux, this uses D-Bus Secret Service.
func Retrieve() ([]byte, error) {
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

// Delete removes the encryption key from the OS keychain.
func Delete() error {
	err := keyring.Delete(serviceName, accountName)
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("failed to delete keychain item: %w", err)
	}

	return nil
}
