package rotate

import (
	"errors"
	"fmt"
	"os"

	"github.com/loderunner/apiki/internal/crypto"
	"github.com/loderunner/apiki/internal/entries"
	"github.com/loderunner/apiki/internal/keychain"
	"github.com/loderunner/apiki/internal/prompt"
)

var ErrNoEntries = errors.New("no variables to re-encrypt")

// Run executes the rotate command.
func Run(path string) error {
	// Load file
	file, err := entries.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load file: %w", err)
	}

	if !file.Encrypted() {
		return errors.New("file is not encrypted")
	}

	if len(file.Entries) == 0 {
		return ErrNoEntries
	}

	oldMode := file.Encryption.Mode
	var oldKey []byte

	// Get old key based on current encryption mode
	switch oldMode {
	case "keychain":
		fmt.Fprintf(os.Stderr, "Unlocking variables with keychain...\n")
		oldKey, err = keychain.Retrieve()
		if err != nil {
			return fmt.Errorf("failed to retrieve key from keychain: %w", err)
		}

	case "password":
		// Prompt for current password until correct
		for {
			password, err := prompt.ReadPassword("Enter current password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			oldKey, err = file.VerifyPassword(password)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Wrong password.\n")
				continue
			}
			break
		}

	default:
		return fmt.Errorf("unknown encryption mode: %q", oldMode)
	}

	// Decrypt all entries with old key
	for i := range file.Entries {
		decrypted, err := crypto.Decrypt(oldKey, file.Entries[i].Value)
		if err != nil {
			return fmt.Errorf(
				"error decrypting variable %q: %w",
				file.Entries[i].Name,
				err,
			)
		}
		file.Entries[i].Value = decrypted
	}

	// Ask for new encryption mode
	newMode, err := prompt.ReadChoice(
		"Lock variables with [p]assword or [k]eychain? ",
		map[rune]string{
			'p': "password",
			'k': "keychain",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to read choice: %w", err)
	}

	var newKey []byte

	switch newMode {
	case "password":
		// Get new password
		password, err := prompt.ReadPassword("Enter new password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		// Confirm password
		passwordConfirm, err := prompt.ReadPassword("Confirm new password: ")
		if err != nil {
			return fmt.Errorf("failed to read password confirmation: %w", err)
		}

		if password != passwordConfirm {
			return errors.New("passwords do not match")
		}

		// Configure new password mode and get the derived key
		newKey, err = file.SetPasswordMode(password)
		if err != nil {
			return fmt.Errorf("failed to configure password mode: %w", err)
		}

	case "keychain":
		// Generate new key
		newKey, err = crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}

		// Delete old keychain entry if it exists
		_ = keychain.Delete()

		// Store new key in keychain
		if err := keychain.Store(newKey); err != nil {
			return fmt.Errorf("failed to store key in keychain: %w", err)
		}

		file.SetKeychainMode()

	default:
		return fmt.Errorf("invalid mode: %q", newMode)
	}

	// Encrypt all entries with new key
	for i := range file.Entries {
		encrypted, err := crypto.Encrypt(newKey, file.Entries[i].Value)
		if err != nil {
			return fmt.Errorf(
				"error encrypting variable %q: %w",
				file.Entries[i].Name,
				err,
			)
		}
		file.Entries[i].Value = encrypted
	}

	// Save file
	if err := entries.Save(path, file); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Fprintf(
		os.Stderr,
		"✓ Re-encrypted %d variables.\n",
		len(file.Entries),
	)

	return nil
}
