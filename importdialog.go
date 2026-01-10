package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateConfirmImport(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "y", "Y", "enter":
		return m.confirmImport()

	case "n", "N", "esc", "q":
		// Return to import mode
		m.mode = modeImport
	}

	return m, nil
}

func (m Model) viewConfirmImport() string {
	var b strings.Builder

	// Count selected entries
	selectedCount := 0
	for _, entry := range m.entries {
		if entry.Selected {
			selectedCount++
		}
	}

	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBrightYellow)
	b.WriteString(warnStyle.Render("Import Variables?"))
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(colorWhite)
	fileStyle := lipgloss.NewStyle().Foreground(colorBrightCyan)

	var variableWord string
	if selectedCount == 1 {
		variableWord = "variable"
	} else {
		variableWord = "variables"
	}

	fmt.Fprintf(&b, "  %s\n",
		infoStyle.Render(fmt.Sprintf(
			"You are about to import %d %s from the current environment to",
			selectedCount,
			variableWord,
		)),
	)
	fmt.Fprintf(&b, "  %s\n\n",
		fileStyle.Render(m.entriesPath),
	)

	return b.String()
}
