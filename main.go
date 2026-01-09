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
		output := generateShellCommands(m.Entries())
		return output, nil
	}

	return "", nil
}

// generateShellCommands produces export and unset statements for the given
// entries. Selected entries get exported, deselected entries get unset.
func generateShellCommands(entries []Entry) string {
	commands := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.Selected {
			// Escape single quotes in value: replace ' with '\''
			escaped := strings.ReplaceAll(entry.Value, "'", "'\\''")
			commands = append(
				commands,
				fmt.Sprintf("export %s='%s'", entry.Name, escaped),
			)
		} else {
			commands = append(commands, fmt.Sprintf("unset %s", entry.Name))
		}
	}

	return strings.Join(commands, "\n")
}
