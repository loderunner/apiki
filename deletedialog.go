package main

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "y", "Y", "enter":
		if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
			actualIndex := m.filteredIndices[m.cursor]
			if actualIndex < len(m.entries) {
				// Save original entries for recovery on persist failure
				originalEntries := slices.Clone(m.entries)

				m.entries = append(
					m.entries[:actualIndex],
					m.entries[actualIndex+1:]...)
				m = m.persistEntries()

				// On persist failure, restore original entries and stay in
				// error mode
				if m.mode == modeError {
					m.entries = originalEntries
					return m, nil
				}

				m = m.recomputeFilter()
				if m.cursor >= len(m.filteredIndices) && m.cursor > 0 {
					m.cursor--
				}
			}
		}
		m.mode = modeList

	case "n", "N", "esc", "q":
		m.mode = modeList
	}

	return m, nil
}

func (m Model) viewConfirmDelete() string {
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

	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBrightYellow)
	b.WriteString(warnStyle.Render("Delete Entry?"))
	b.WriteString("\n\n")

	nameStyle := lipgloss.NewStyle().Bold(true)
	fmt.Fprintf(&b, "  %s", nameStyle.Render(entry.Name))
	if entry.Label != "" {
		labelStyle := lipgloss.NewStyle().Foreground(colorGray).Italic(true)
		fmt.Fprintf(&b, " %s", labelStyle.Render(entry.Label))
	}
	b.WriteString("\n\n")

	return b.String()
}
