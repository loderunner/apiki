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

	// Sync selection state with current environment
	SyncWithEnvironment(entries)

	// Capture which variables were originally set in the environment
	originallySet := make(map[string]bool)
	for _, entry := range entries {
		originallySet[entry.Name] = entry.Selected
	}

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
		output := generateShellCommands(m.Entries(), originallySet)
		return output, nil
	}

	return "", nil
}

// generateShellCommands produces export and unset statements for the given
// entries. Selected entries get exported. Only entries that were originally
// set in the environment get unset when deselected.
func generateShellCommands(
	entries []Entry,
	originallySet map[string]bool,
) string {
	commands := make([]string, 0, len(entries))
	handledNames := make(map[string]struct{})

	for _, entry := range entries {
		if _, ok := handledNames[entry.Name]; ok {
			continue
		}
		handledNames[entry.Name] = struct{}{}

		if entry.Selected {
			// Escape single quotes in value: replace ' with '\''
			escaped := strings.ReplaceAll(entry.Value, "'", "'\\''")
			commands = append(
				commands,
				fmt.Sprintf("export %s='%s'", entry.Name, escaped),
			)
		} else if originallySet[entry.Name] {
			// Only unset variables that were originally set
			commands = append(commands, fmt.Sprintf("unset %s", entry.Name))
		}
	}

	return strings.Join(commands, "\n")
}
