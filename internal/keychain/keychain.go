package keychain

import (
	"fmt"

	"github.com/keybase/go-keychain"
)

const (
	serviceName = "apiki"
	accountName = "encryption-key"
	label       = "apiki encryption key"
	accessGroup = "apiki"
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

	item := keychain.NewGenericPassword(
		serviceName,
		accountName,
		label,
		key,
		accessGroup,
	)

	// macOS: SetAccessible to require authentication when unlocked
	// passcode prompt happens automatically on macOS when retrieving
	item.SetAccessible(keychain.AccessibleAfterFirstUnlockThisDeviceOnly)

	err := keychain.AddItem(item)
	if err != nil {
		// If item already exists, delete and re-add
		if err == keychain.ErrorDuplicateItem {
			if delErr := Delete(); delErr != nil {
				return fmt.Errorf(
					"failed to delete existing keychain item: %w",
					delErr,
				)
			}
			err = keychain.AddItem(item)
		}
		if err != nil {
			return fmt.Errorf("failed to store key in keychain: %w", err)
		}
	}

	return nil
}

// Retrieve retrieves the encryption key from the OS keychain.
// On macOS, this uses the macOS Keychain API.
// On Linux, this uses D-Bus Secret Service.
func Retrieve() ([]byte, error) {
	query := keychain.NewGenericPassword(
		serviceName,
		accountName,
		label,
		nil,
		accessGroup,
	)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query keychain: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("key not found in keychain")
	}

	key := results[0].Data
	if len(key) != 32 {
		return nil, fmt.Errorf(
			"invalid key size in keychain: expected 32 bytes, got %d",
			len(key),
		)
	}

	return key, nil
}

// Delete removes the encryption key from the OS keychain.
func Delete() error {
	query := keychain.NewGenericPassword(
		serviceName,
		accountName,
		label,
		nil,
		accessGroup,
	)

	err := keychain.DeleteItem(query)
	if err != nil && err != keychain.ErrorItemNotFound {
		return fmt.Errorf("failed to delete keychain item: %w", err)
	}

	return nil
}
