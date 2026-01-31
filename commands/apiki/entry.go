package apiki

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/loderunner/apiki/internal/entries"
)

// Entry represents an environment variable entry managed by apiki.
type Entry struct {
	entries.Entry

	// Selected indicates whether this entry should be exported (true) or unset
	// (false).
	Selected bool

	// SourceFile is the path to the .env file this entry came from.
	// Empty string means the entry came from the apiki file.
	SourceFile string
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
		return strings.Compare(
			strings.ToLower(a.Label),
			strings.ToLower(b.Label),
		)
	})
}
