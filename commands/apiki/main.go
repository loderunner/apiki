package apiki

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/loderunner/apiki/commands"
	"github.com/loderunner/apiki/internal/entries"
)

// Run executes the apiki root command
func Run(variablesPath, configPath string) (string, error) {
	// Load file (may be encrypted)
	file, err := entries.Load(variablesPath)
	if err != nil {
		return "", fmt.Errorf("could not load variables file: %w", err)
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

	// Convert entries.File entries to TUI Entry format
	apikiEntries := make([]Entry, len(file.Entries))
	for i, e := range file.Entries {
		apikiEntries[i] = Entry{
			Entry:      e,
			Selected:   false,
			SourceFile: "",
		}
	}

	dotEnvEntries, err := LoadDotEnvEntries()
	if err != nil {
		return "", fmt.Errorf("could not load .env variables: %w", err)
	}

	// Combine apiki entries with .env entries (no deduplication)
	allEntries := append(apikiEntries, dotEnvEntries...)

	// Sort all entries together
	SortEntries(allEntries)

	// Capture the environment state for all entry names at startup
	envSnapshot := captureEnvironment(allEntries)

	// Sync selection state with captured environment (for all entries)
	syncWithEnvironment(allEntries, envSnapshot)

	// Open TTY for TUI input/output, keeping stdout clean for shell commands
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("could not open /dev/tty: %w", err)
	}
	defer func() { _ = tty.Close() }()

	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(tty))

	model := NewModel(
		file,
		variablesPath,
		configPath,
		encryptionKey,
		allEntries,
	)
	p := tea.NewProgram(model, tea.WithInput(tty), tea.WithOutput(tty))

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	m, ok := finalModel.(Model)
	if !ok {
		panic("unexpected error")
	}

	// Clear the TUI from the terminal
	viewOutput := m.View()
	lineCount := strings.Count(viewOutput, "\n")
	clearSequence := fmt.Sprintf("\033[%dA\033[J", lineCount)
	_, _ = tty.WriteString(clearSequence)

	// If cancelled (Ctrl-C), exit without saving or outputting
	if m.Cancelled() {
		return "", nil
	}

	// If quitting normally, save entries and output shell commands
	if m.Quitting() {
		// Save is handled by persistEntries in the model
		// Output export/unset commands to stdout
		output := generateShellCommands(m.Entries(), envSnapshot)
		return output, nil
	}

	return "", nil
}

// captureEnvironment builds a snapshot of environment variable values for all
// entry names. Returns a map of name -> value (empty string if not set).
func captureEnvironment(entries []Entry) map[string]string {
	env := make(map[string]string)
	for _, entry := range entries {
		if _, ok := env[entry.Name]; !ok {
			env[entry.Name] = os.Getenv(entry.Name)
		}
	}
	return env
}

// syncWithEnvironment updates the Selected state of each variable based on the
// captured environment snapshot.
//
// A variable is marked as Selected if both its Name and Value match the
// environment. For variables with duplicate names (radio groups), only the
// variable whose value matches the environment is selected. If no exact match
// is found, no variable with that name is selected.
func syncWithEnvironment(entries []Entry, env map[string]string) {
	selectedNames := make(map[string]struct{})

	for i := range entries {
		name := entries[i].Name
		envVal := env[name]

		if envVal == "" {
			entries[i].Selected = false
			continue
		}

		// If we already selected a variable for this name, skip
		if _, ok := selectedNames[name]; ok {
			entries[i].Selected = false
			continue
		}

		// Select only if both name and value match the environment
		if entries[i].Value == envVal {
			entries[i].Selected = true
			selectedNames[name] = struct{}{}
		} else {
			entries[i].Selected = false
		}
	}
}

// generateShellCommands produces export and unset statements for the given
// variables. Only outputs commands when the value has actually changed from the
// original environment state.
func generateShellCommands(entries []Entry, env map[string]string) string {
	// Build a map of name -> selected variable (if any) for radio-group
	// handling
	selectedByName := make(map[string]*Entry)
	for i := range entries {
		if entries[i].Selected {
			selectedByName[entries[i].Name] = &entries[i]
		}
	}

	commands := make([]string, 0, len(entries))
	handledNames := make(map[string]struct{})

	for _, entry := range entries {
		if _, ok := handledNames[entry.Name]; ok {
			continue
		}
		handledNames[entry.Name] = struct{}{}

		originalValue := env[entry.Name]

		if selected, ok := selectedByName[entry.Name]; ok {
			// Only export if the value differs from the original
			if selected.Value != originalValue {
				escaped := strings.ReplaceAll(selected.Value, "'", "'\\''")
				commands = append(
					commands,
					fmt.Sprintf("export %s='%s'", selected.Name, escaped),
				)
			}
		} else if originalValue != "" {
			// No variable with this name is selected, unset if it was
			// originally
			// set
			commands = append(commands, fmt.Sprintf("unset %s", entry.Name))
		}
	}

	return strings.Join(commands, "\n")
}
