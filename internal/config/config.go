package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/spf13/afero"

	"github.com/loderunner/apiki/internal/entries"
	"github.com/loderunner/apiki/internal/set"
)

var fs = afero.NewOsFs()

// Config represents the apiki configuration file.
type Config struct {
	Selected set.Set[string] `json:"selected,omitempty"`
}

// Load reads the config file from disk and parses it into memory.
// Returns an empty config if the file doesn't exist.
func Load(path string) (*Config, error) {
	dir := filepath.Dir(path)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := afero.ReadFile(fs, path)
	if err != nil {
		if errors.Is(err, afero.ErrFileNotFound) {
			return &Config{
				Selected: set.New[string](),
			}, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return &Config{
			Selected: set.New[string](),
		}, nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Initialize Selected if nil (for backward compatibility)
	if cfg.Selected == nil {
		cfg.Selected = set.New[string]()
	}

	return &cfg, nil
}

// Save serializes the config and writes it to disk.
func Save(path string, c *Config) error {
	dir := filepath.Dir(path)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := afero.WriteFile(fs, path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// EntryID computes the unique identifier for an entry at the given index.
// For entries with unique names, returns just the name.
// For entries in radio groups (same name), returns "name[index]" where index
// is the position within the radio group after sorting.
func EntryID(entries []entries.Entry, index int) string {
	if index < 0 || index >= len(entries) {
		return ""
	}

	entry := entries[index]
	name := entry.Name

	// Count how many entries have the same name
	nameCount := 0
	for _, e := range entries {
		if e.Name == name {
			nameCount++
		}
	}

	// If unique, return just the name
	if nameCount == 1 {
		return name
	}

	// For radio groups, find the index within entries with the same name
	// Entries are sorted, so we need to find position within the group
	sameNameIndices := make([]int, 0)
	for i, e := range entries {
		if e.Name == name {
			sameNameIndices = append(sameNameIndices, i)
		}
	}

	// Find position of current index within same-name entries
	pos := slices.Index(sameNameIndices, index)
	if pos == -1 {
		return name // fallback
	}

	return fmt.Sprintf("%s[%d]", name, pos)
}
