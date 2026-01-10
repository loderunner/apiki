package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateError(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.errorMessage = ""
		m.mode = modeList
	case "q", "ctrl+c":
		m.cancelled = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) viewError() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorBrightBlue)
	b.WriteString(titleStyle.Render("Error"))
	b.WriteString("\n\n")

	errorStyle := lipgloss.NewStyle().
		Foreground(colorBrightRed)
	b.WriteString("  ")
	b.WriteString(errorStyle.Render(m.errorMessage))
	b.WriteString("\n")

	return b.String()
}
