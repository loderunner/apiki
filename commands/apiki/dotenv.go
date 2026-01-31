package apiki

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"github.com/loderunner/apiki/internal/entries"
)

// FindDotEnvFiles walks upward from startDir and collects all .env and .env.*
// files.
//
// Returns a slice of absolute file paths, ordered from deepest to shallowest.
func FindDotEnvFiles(startDir string) ([]string, error) {
	var files []string
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}

	for {
		entries, err := os.ReadDir(dir)
		if err != nil {
			// If we can't read the directory, stop searching upward
			break
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if name == ".env" || strings.HasPrefix(name, ".env.") {
				fullPath := filepath.Join(dir, name)
				files = append(files, fullPath)
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return files, nil
}

// ParseDotEnvFile parses a single .env file and converts it to Entry slice.
// Each entry gets a label of the form "from <dirname>/<filename>".
func ParseDotEnvFile(path string) ([]Entry, error) {
	envMap, err := godotenv.Read(path)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	dirname := filepath.Base(dir)
	label := "from " + dirname + "/" + filename

	result := make([]Entry, 0, len(envMap))
	for name, value := range envMap {
		result = append(result, Entry{
			Entry: entries.Entry{
				Name:  name,
				Value: value,
				Label: label,
			},
			Selected:   false,
			SourceFile: path,
		})
	}

	return result, nil
}

// LoadDotEnvEntries finds and parses all .env files upward from PWD.
// Returns all entries from all found .env files.
func LoadDotEnvEntries() ([]Entry, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	envFiles, err := FindDotEnvFiles(pwd)
	if err != nil {
		return nil, err
	}

	var allEntries []Entry
	for _, envFile := range envFiles {
		entries, err := ParseDotEnvFile(envFile)
		if err != nil {
			// Skip files that can't be parsed, but continue with others
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}
