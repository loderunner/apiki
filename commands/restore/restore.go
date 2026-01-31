package restore

import (
	"fmt"
	"strings"

	"github.com/loderunner/apiki/commands"
	"github.com/loderunner/apiki/internal/config"
	"github.com/loderunner/apiki/internal/entries"
)

// Run loads the config and variables files, then outputs export commands for
// selected entries. Returns empty string if no entries are selected.
func Run(variablesPath, configPath string) (string, error) {
	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		return "", fmt.Errorf("could not load config file: %w", err)
	}

	// Load variables file
	file, err := entries.Load(variablesPath)
	if err != nil {
		return "", fmt.Errorf("could not load variables file: %w", err)
	}

	// If file is empty, return empty output
	if len(file.Entries) == 0 {
		return "", nil
	}

	// Unlock if encrypted
	var encryptionKey []byte
	if file.Encrypted() {
		encryptionKey, err = commands.Unlock(file)
		if err != nil {
			return "", fmt.Errorf("failed to unlock file: %w", err)
		}

		// Decrypt values in memory
		if err := file.DecryptValues(encryptionKey); err != nil {
			return "", fmt.Errorf("failed to decrypt variables: %w", err)
		}
	}

	// Generate export commands for selected entries
	var commands []string
	for i, entry := range file.Entries {
		entryID := config.EntryID(file.Entries, i)
		if cfg.Selected.Has(entryID) {
			escaped := strings.ReplaceAll(entry.Value, "'", "'\\''")
			commands = append(
				commands,
				fmt.Sprintf("export %s='%s'", entry.Name, escaped),
			)
		}
	}

	return strings.Join(commands, "\n"), nil
}
