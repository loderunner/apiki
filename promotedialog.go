package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateConfirmPromote(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "y", "Y", "enter":
		if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
			actualIndex := m.filteredIndices[m.cursor]
			if actualIndex < len(m.entries) {
				entry := m.entries[actualIndex]
				// Proceed to edit form with the entry values pre-filled
				// editIndex = -1 will create a new entry
				m.mode = modeEdit
				var cmd tea.Cmd
				m, cmd = m.prepareForm(-1, &entry)
				return m, cmd
			}
		}
		m.mode = modeList

	case "n", "N", "esc", "q":
		m.mode = modeList
	}

	return m, nil
}

func (m Model) viewConfirmPromote() string {
	var b strings.Builder

	noFiltered := len(m.filteredIndices) == 0
	cursorOutOfBounds := m.cursor < 0 || m.cursor >= len(m.filteredIndices)
	if noFiltered || cursorOutOfBounds {
		return ""
	}

	actualIndex := m.filteredIndices[m.cursor]
	if actualIndex < 0 || actualIndex >= len(m.entries) {
		return ""
	}

	entry := m.entries[actualIndex]

	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorBrightYellow)
	b.WriteString(warnStyle.Render("Add to apiki?"))
	b.WriteString("\n\n")

	nameStyle := lipgloss.NewStyle().Bold(true)
	fmt.Fprintf(&b, "  %s", nameStyle.Render(entry.Name))
	if entry.Label != "" {
		labelStyle := lipgloss.NewStyle().Foreground(ColorGray).Italic(true)
		fmt.Fprintf(&b, " %s", labelStyle.Render(entry.Label))
	}
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(ColorGray)
	b.WriteString(
		infoStyle.Render(
			"  This will create a new variable in your apiki file.\n",
		),
	)

	return b.String()
}
