package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Entry represents an environment variable entry managed by apiki.
type Entry struct {
	// Name is the environment variable name (e.g., "PATH", "DATABASE_URL").
	Name string `json:"name"`

	// Value is the value to set when the entry is selected.
	Value string `json:"value"`

	// Label is a human-readable description of this entry.
	Label string `json:"label"`

	// Selected indicates whether this entry should be exported (true) or unset
	// (false). Not serialized to JSON.
	Selected bool `json:"-"`

	// SourceFile is the path to the .env file this entry came from.
	// Empty string means the entry came from the apiki file.
	// Not serialized to JSON.
	SourceFile string `json:"-"`
}

// FuzzyTarget returns the string to use for fuzzy matching this entry.
// For .env entries, this includes the directory/filename without the "from"
// prefix.
func (e Entry) FuzzyTarget() string {
	if e.SourceFile == "" {
		if e.Label == "" {
			return e.Name
		}
		return e.Name + " " + e.Label
	}

	// Extract dirname/filename from SourceFile for search
	// SourceFile contains the full path, but we want just dirname/filename
	dir := filepath.Dir(e.SourceFile)
	filename := filepath.Base(e.SourceFile)
	dirname := filepath.Base(dir)
	return e.Name + " " + dirname + "/" + filename
}

// SortEntries sorts entries alphabetically by (Name, Label), case-insensitive.
func SortEntries(entries []Entry) {
	slices.SortFunc(entries, func(a, b Entry) int {
		if c := strings.Compare(
			strings.ToLower(a.Name),
			strings.ToLower(b.Name),
		); c != 0 {
			return c
		}
		return strings.Compare(strings.ToLower(a.Label), strings.ToLower(b.Label))
	})
}

// DefaultEntriesPath returns the default path for the entries file:
// ~/.apiki/variables.json
func DefaultEntriesPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".apiki", "variables.json"), nil
}

// LoadEntries reads entries from the given JSON file path.
//
// If the file does not exist, it returns an empty slice (not an error).
// The directory is created if it does not exist.
func LoadEntries(path string) ([]Entry, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []Entry{}, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return []Entry{}, nil
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// SaveEntries writes entries to the given JSON file path.
// The file with full path is created if it does not exist.
func SaveEntries(path string, entries []Entry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
