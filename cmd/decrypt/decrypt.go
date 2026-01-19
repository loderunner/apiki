package decrypt

import (
	"errors"
	"fmt"
	"os"

	"github.com/loderunner/apiki/internal/crypto"
	"github.com/loderunner/apiki/internal/entries"
	"github.com/loderunner/apiki/internal/keychain"
	"github.com/loderunner/apiki/internal/prompt"
)

var ErrNoEntries = errors.New("no variables to decrypt")

// Run executes the decrypt command.
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

	var key []byte

	switch file.Encryption.Mode {
	case "keychain":
		// Retrieve key from keychain
		fmt.Fprintf(os.Stderr, "Unlocking variables with keychain...\n")
		key, err = keychain.Retrieve()
		if err != nil {
			return fmt.Errorf("failed to retrieve key from keychain: %w", err)
		}

	case "password":
		// Prompt for password until correct
		for {
			password, err := prompt.ReadPassword("Enter password: ")
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}

			key, err = file.VerifyPassword(password)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Wrong password.\n")
				continue
			}
			break
		}

	default:
		return fmt.Errorf("unknown encryption mode: %q", file.Encryption.Mode)
	}

	// Ask for confirmation
	confirm, err := prompt.ReadChoiceWithDefault(
		"Values will be stored in plaintext. Continue? [Y/n] ",
		map[rune]string{
			'y': "yes",
			'n': "no",
		},
		"yes",
	)
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	if confirm == "no" {
		return nil
	}

	// Decrypt all entries
	for i := range file.Entries {
		decrypted, err := crypto.Decrypt(key, file.Entries[i].Value)
		if err != nil {
			return fmt.Errorf(
				"error decrypting variable %q: %w",
				file.Entries[i].Name,
				err,
			)
		}
		file.Entries[i].Value = decrypted
	}

	// Clear encryption header
	file.ClearEncryption()

	// Save file
	if err := entries.Save(path, file); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Fprintf(
		os.Stderr,
		"✓ Decrypted %d variables. Values are now stored in plaintext.\n",
		len(file.Entries),
	)

	return nil
}
