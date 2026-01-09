package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// nameGroups returns a map of name -> indices of entries with that name.
// Used to identify entries that belong to the same group (same variable name).
func (m Model) nameGroups() map[string][]int {
	if m.nameGroupsMemo != nil {
		return m.nameGroupsMemo
	}

	groups := make(map[string][]int)
	for i, entry := range m.entries {
		groups[entry.Name] = append(groups[entry.Name], i)
	}
	m.nameGroupsMemo = groups
	return m.nameGroupsMemo
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Filter input mode: only handle esc/enter, pass everything else to input
	if m.filtering {
		switch key {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			m = m.clearFilter()
			return m, nil
		case "enter":
			m.filtering = false
			m.filterInput.Blur()
			m = m.recomputeFilter()
			return m, nil
		}
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m = m.recomputeFilter()
		return m, cmd
	}

	// List mode keys
	switch key {
	case "ctrl+c", "q":
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		if m.filterInput.Value() != "" {
			m = m.clearFilter()
			return m, nil
		}
		m.cancelled = true
		return m, tea.Quit

	case "enter":
		m.quitting = true
		return m, tea.Quit

	case "/":
		m.filtering = true
		m.filterInput.SetValue("")
		m.filterInput.Focus()
		m = m.recomputeFilter()
		return m, textinput.Blink

	case "up", "k":
		if len(m.filteredIndices) == 0 {
			return m, nil
		}
		if m.cursor > 0 {
			m.cursor--
		} else {
			m.cursor = len(m.filteredIndices) - 1
		}
		return m, nil

	case "down", "j":
		if len(m.filteredIndices) == 0 {
			return m, nil
		}
		if m.cursor < len(m.filteredIndices)-1 {
			m.cursor++
		} else {
			m.cursor = 0
		}
		return m, nil

	case "=":
		if len(m.filteredIndices) > 0 {
			m = m.clearFilter()
			actualIndex := m.filteredIndices[m.cursor]
			m.mode = modeEdit
			m.editIndex = actualIndex
			entry := m.entries[actualIndex]
			m.nameInput.SetValue(entry.Name)
			m.valueInput.SetValue(entry.Value)
			m.labelInput.SetValue(entry.Label)
			m.currentField = fieldName
			m.nameError = ""
			m.valueError = ""
			m = m.updateInputWidths()
			m.nameInput.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case " ":
		if len(m.filteredIndices) > 0 {
			actualIndex := m.filteredIndices[m.cursor]
			currentEntry := &m.entries[actualIndex]
			currentEntry.Selected = !currentEntry.Selected

			// Radio-button behavior: if we selected this entry, deselect others
			// with the same name
			if currentEntry.Selected {
				groups := m.nameGroups()
				for _, i := range groups[currentEntry.Name] {
					if i != actualIndex {
						m.entries[i].Selected = false
					}
				}
			}
		}

	case "+":
		m.clearFilter()
		m.mode = modeAdd
		m.editIndex = -1
		m.nameInput.SetValue("")
		m.valueInput.SetValue("")
		m.labelInput.SetValue("")
		m.currentField = fieldName
		m.nameError = ""
		m.valueError = ""
		m = m.updateInputWidths()
		m.nameInput.Focus()
		return m, textinput.Blink

	case "backspace", "delete", "-":
		if len(m.filteredIndices) > 0 {
			m.mode = modeConfirmDelete
		}
	}

	return m, nil
}

func (m Model) viewList() string {
	var b strings.Builder

	titleStyle := lipgloss.
		NewStyle().
		Bold(true).
		Foreground(colorBrightBlue)
	b.WriteString(titleStyle.Render("Environment Variables"))
	b.WriteString("\n\n")

	dimStyle := lipgloss.NewStyle().Foreground(colorGray)
	if len(m.entries) == 0 {
		b.WriteString(dimStyle.Render("  No entries. Press + to add one."))
		b.WriteString("\n")
		return b.String()
	}

	selectedStyle := lipgloss.
		NewStyle().
		Foreground(colorBrightGreen)
	unselectedStyle := lipgloss.
		NewStyle().
		Foreground(colorGray)
	cursorStyle := lipgloss.NewStyle().Bold(true)
	nameStyle := lipgloss.NewStyle().Bold(true)
	labelStyle := lipgloss.
		NewStyle().
		Foreground(colorGray).Italic(true)
	groupConnectorStyle := lipgloss.
		NewStyle().
		Foreground(colorGray)

	groups := m.nameGroups()

	entriesToShow := m.filteredIndices
	if len(entriesToShow) == 0 {
		b.WriteString(dimStyle.Render("  No entries match the filter."))
		b.WriteString("\n")
		return b.String()
	}

	for displayIdx, actualIdx := range entriesToShow {
		entry := m.entries[actualIdx]
		cursor := "  "
		if displayIdx == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		var checkbox string
		if entry.Selected {
			checkbox = selectedStyle.Render("⦿ ")
		} else {
			checkbox = unselectedStyle.Render("◯ ")
		}

		// Check if this entry is part of a group (multiple entries with same name)
		groupIndices := groups[entry.Name]
		grouped := len(groupIndices) > 1

		var groupPrefix string
		if grouped {
			// Find position within group
			posInGroup := 0
			for j, idx := range groupIndices {
				if idx == actualIdx {
					posInGroup = j
					break
				}
			}

			if posInGroup == 0 {
				// First in group - show corner
				groupPrefix = groupConnectorStyle.Render("┌ ")
			} else if posInGroup == len(groupIndices)-1 {
				// Last in group
				groupPrefix = groupConnectorStyle.Render("└ ")
			} else {
				// Middle of group
				groupPrefix = groupConnectorStyle.Render("├ ")
			}
		} else {
			groupPrefix = "  "
		}

		name := nameStyle.Render(entry.Name)
		label := ""
		if entry.Label != "" {
			label = " " + labelStyle.Render(entry.Label)
		}

		fmt.Fprintf(&b, "%s%s%s%s%s\n", cursor, groupPrefix, checkbox, name, label)
	}

	return b.String()
}
