package commands

import (
	"fmt"
	"os"

	"github.com/loderunner/apiki/internal/entries"
	"github.com/loderunner/apiki/internal/keychain"
	"github.com/loderunner/apiki/internal/prompt"
)

// Unlock prompts for password or retrieves key from keychain to unlock
// an encrypted file. Returns the encryption key.
func Unlock(file *entries.File) ([]byte, error) {
	if !file.Encrypted() {
		return nil, fmt.Errorf("file is not encrypted")
	}

	if file.Encryption.Mode == "password" {
		// Check for APIKI_PASSWORD environment variable
		if password := os.Getenv("APIKI_PASSWORD"); password != "" {
			key, err := file.VerifyPassword(password)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid password from APIKI_PASSWORD: %w",
					err,
				)
			}
			return key, nil
		}

		// Prompt for password
		firstAttempt := true
		for {
			password, err := prompt.ReadPassword("Enter password: ")
			if err != nil {
				return nil, fmt.Errorf("failed to read password: %w", err)
			}

			key, err := file.VerifyPassword(password)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Wrong password.\n")
				if firstAttempt {
					firstAttempt = false
					continue
				}

				return nil, fmt.Errorf("too many wrong password attempts")
			}
			return key, nil
		}
	} else if file.Encryption.Mode == "keychain" {
		// Retrieve from keychain (may trigger Touch ID on macOS)
		fmt.Fprintf(os.Stderr, "Unlocking variables with keychain...\n")
		key, err := keychain.Retrieve()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to retrieve key from keychain: %w",
				err,
			)
		}
		return key, nil
	}

	return nil, fmt.Errorf("unknown encryption mode: %q", file.Encryption.Mode)
}
