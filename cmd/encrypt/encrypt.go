package encrypt

import (
	"errors"
	"fmt"
	"os"

	"github.com/loderunner/apiki/internal/crypto"
	"github.com/loderunner/apiki/internal/entries"
	"github.com/loderunner/apiki/internal/keychain"
	"github.com/loderunner/apiki/internal/prompt"
)

var ErrNoEntries = errors.New("no variables to encrypt")

// Run executes the encrypt command.
func Run(path string) error {
	// Load file
	file, err := entries.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load file: %w", err)
	}

	if file.Encrypted() {
		return errors.New(
			"file is already encrypted, " +
				"use `apiki rotate` to rotate the encryption key",
		)
	}

	if len(file.Entries) == 0 {
		return ErrNoEntries
	}

	// Ask for encryption mode
	mode, err := prompt.ReadChoice(
		"Lock variables with [p]assword or [k]eychain? ",
		map[rune]string{
			'p': "password",
			'k': "keychain",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to read choice: %w", err)
	}

	var key []byte

	switch mode {
	case "password":
		// Get password
		password, err := prompt.ReadPassword("Enter password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		// Confirm password
		passwordConfirm, err := prompt.ReadPassword("Confirm password: ")
		if err != nil {
			return fmt.Errorf("failed to read password confirmation: %w", err)
		}

		if password != passwordConfirm {
			return errors.New("passwords do not match")
		}

		// Configure password mode and get the derived key
		key, err = file.SetPasswordMode(password)
		if err != nil {
			return fmt.Errorf(
				"failed to configure variables file "+
					"for password encryption: %w",
				err,
			)
		}

	case "keychain":
		// Generate key and store in keychain
		key, err = crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}

		if err := keychain.Store(key); err != nil {
			return fmt.Errorf("failed to store key in keychain: %w", err)
		}

		file.SetKeychainMode()

	default:
		return fmt.Errorf("invalid unlock method: %q", mode)
	}

	// Encrypt all entries
	for i := range file.Entries {
		encrypted, err := crypto.Encrypt(key, file.Entries[i].Value)
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
		"✓ Encrypted %d variables.\n",
		len(file.Entries),
	)

	return nil
}
