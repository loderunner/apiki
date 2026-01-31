package apiki

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/loderunner/apiki/internal/entries"
)

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		m = m.clearFilter()
		m.mode = modeList
		m.nameError = ""
		m.valueError = ""
		return m, nil

	case "tab", "down":
		return m.nextField()

	case "shift+tab", "up":
		return m.prevField()

	case "enter":
		if m.currentField == fieldLabel {
			return m.saveFormEntry()
		}
		return m.nextField()
	}

	// Update the focused input
	var cmd tea.Cmd
	switch m.currentField {
	case fieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
		// Clear error when user starts typing
		if m.nameError != "" {
			m.nameError = ""
		}
	case fieldValue:
		m.valueInput, cmd = m.valueInput.Update(msg)
		// Clear error when user starts typing
		if m.valueError != "" {
			m.valueError = ""
		}
	case fieldLabel:
		m.labelInput, cmd = m.labelInput.Update(msg)
	}

	m = m.updateInputWidths()
	return m, cmd
}

func (m Model) nextField() (tea.Model, tea.Cmd) {
	m.nameInput.Blur()
	m.valueInput.Blur()
	m.labelInput.Blur()

	switch m.currentField {
	case fieldName:
		m.currentField = fieldValue
		m.valueInput.Focus()
	case fieldValue:
		m.currentField = fieldLabel
		m.labelInput.Focus()
	case fieldLabel:
		m.currentField = fieldName
		m.nameInput.Focus()
	}

	return m, textinput.Blink
}

func (m Model) prevField() (tea.Model, tea.Cmd) {
	m.nameInput.Blur()
	m.valueInput.Blur()
	m.labelInput.Blur()

	switch m.currentField {
	case fieldName:
		m.currentField = fieldLabel
		m.labelInput.Focus()
	case fieldValue:
		m.currentField = fieldName
		m.nameInput.Focus()
	case fieldLabel:
		m.currentField = fieldValue
		m.valueInput.Focus()
	}

	return m, textinput.Blink
}

func (m Model) saveFormEntry() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	value := m.valueInput.Value()

	// Validate fields
	m.nameError = ""
	m.valueError = ""
	valid := true

	if name == "" {
		m.nameError = "name cannot be empty"
		valid = false
	}

	if value == "" {
		m.valueError = "value cannot be empty"
		valid = false
	}

	// Don't save if validation fails
	if !valid {
		return m, nil
	}

	entry := Entry{
		Entry: entries.Entry{
			Name:  name,
			Value: value,
			Label: strings.TrimSpace(m.labelInput.Value()),
		},
		Selected: false,
	}

	// Save original entries for recovery on persist failure
	originalEntries := slices.Clone(m.entries)

	if m.editIndex >= 0 {
		if m.editIndex < len(m.entries) {
			// Preserve selection state when editing
			entry.Selected = m.entries[m.editIndex].Selected
			m.entries[m.editIndex] = entry
		}
	} else {
		m.entries = append(m.entries, entry)
	}

	SortEntries(m.entries)
	m = m.persistEntries()

	// On persist failure, restore original entries and stay in error mode
	if m.mode == modeError {
		m.entries = originalEntries
		return m, nil
	}

	m = m.clearFilter()

	// Move cursor to the saved entry's new position
	for i, e := range m.entries {
		if e.Name == entry.Name && e.Label == entry.Label &&
			e.Value == entry.Value {
			m.cursor = i
			break
		}
	}

	// Adjust cursor to filtered view if needed
	if len(m.filteredIndices) > 0 {
		for displayIdx, actualIdx := range m.filteredIndices {
			if actualIdx == m.cursor {
				m.cursor = displayIdx
				break
			}
		}
		// If cursor entry not in filtered results, reset to 0
		if m.cursor >= len(m.filteredIndices) {
			m.cursor = 0
		}
	}

	m.mode = modeList
	m.nameError = ""
	m.valueError = ""
	return m, nil
}

func (m Model) viewForm(title string) string {
	var b strings.Builder

	titleStyle := lipgloss.
		NewStyle().
		Bold(true).
		Foreground(ColorBrightBlue)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().Width(8)
	errorStyle := lipgloss.
		NewStyle().
		Foreground(ColorBrightRed).
		Italic(true)

	b.WriteString(labelStyle.Render("Name:"))
	b.WriteString(m.nameInput.View())
	if m.nameError != "" {
		b.WriteString(" ")
		b.WriteString(errorStyle.Render(m.nameError))
	}
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Value:"))
	b.WriteString(m.valueInput.View())
	if m.valueError != "" {
		b.WriteString(" ")
		b.WriteString(errorStyle.Render(m.valueError))
	}
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Label:"))
	b.WriteString(m.labelInput.View())
	b.WriteString("\n")

	return b.String()
}
