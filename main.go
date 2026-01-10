package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	output, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if output != "" {
		fmt.Printf("%s\n", output)
	}
}

func run() (string, error) {
	entriesPath, err := DefaultEntriesPath()
	if err != nil {
		return "", fmt.Errorf("could not get entries path: %w", err)
	}

	entries, err := LoadEntries(entriesPath)
	if err != nil {
		return "", fmt.Errorf("could not load entries: %w", err)
	}

	// Capture the environment state for all entry names at startup
	envSnapshot := captureEnvironment(entries)

	// Sync selection state with captured environment
	syncWithEnvironment(entries, envSnapshot)

	// Open TTY for TUI input/output, keeping stdout clean for shell commands
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("could not open /dev/tty: %w", err)
	}
	defer func() { _ = tty.Close() }()

	model := NewModel(entries)
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
		if err := SaveEntries(entriesPath, m.Entries()); err != nil {
			return "", fmt.Errorf("could not save apiki entries file: %w", err)
		}

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

// syncWithEnvironment updates the Selected state of each entry based on the
// captured environment snapshot.
//
// An entry is marked as Selected if both its Name and Value match the
// environment. For entries with duplicate names (radio groups), only the entry
// whose value matches the environment is selected. If no exact match is found,
// no entry with that name is selected.
func syncWithEnvironment(entries []Entry, env map[string]string) {
	selectedNames := make(map[string]struct{})

	for i := range entries {
		name := entries[i].Name
		envVal := env[name]

		if envVal == "" {
			entries[i].Selected = false
			continue
		}

		// If we already selected an entry for this name, skip
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
// entries. Only outputs commands when the value has actually changed from the
// original environment state.
func generateShellCommands(entries []Entry, env map[string]string) string {
	// Build a map of name -> selected entry (if any) for radio-group handling
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
			// No entry with this name is selected, unset if it was originally set
			commands = append(commands, fmt.Sprintf("unset %s", entry.Name))
		}
	}

	return strings.Join(commands, "\n")
}
