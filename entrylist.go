package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// highlightMatches highlights matching characters in the text portion using
// precomputed match positions from fuzzyMatches. The offset parameter indicates
// the starting position of text within the target string used for fuzzy
// matching.
func highlightMatches(
	matchedIndexes []int,
	text string,
	offset int,
	baseStyle lipgloss.Style,
) string {
	if len(matchedIndexes) == 0 {
		return baseStyle.Render(text)
	}

	// Convert matched indexes to a map for fast lookup
	matchPositions := make(map[int]bool)
	for _, idx := range matchedIndexes {
		matchPositions[idx] = true
	}

	// Highlight matching characters in text
	highlightStyle := baseStyle.
		Foreground(colorBrightYellow).
		Bold(true)

	var result strings.Builder
	for i, r := range text {
		pos := offset + i
		if matchPositions[pos] {
			result.WriteString(highlightStyle.Render(string(r)))
		} else {
			result.WriteString(baseStyle.Render(string(r)))
		}
	}

	return result.String()
}

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

// listHeight calculates how many entries can fit in the visible area.
func (m Model) listHeight() int {
	// Fixed overhead: title(1) + top spacer(1) + bottom line(1) + helpbar(1)
	// The bottom line contains ▼ chevron and/or filter bar (they share the line)
	overhead := 4
	visible := m.height - overhead
	if visible < 1 {
		return 1
	}
	return visible
}

// adjustViewport ensures the cursor stays within the visible viewport.
// The viewport only scrolls when the cursor would move outside the visible
// range.
func (m Model) adjustViewport() Model {
	if len(m.filteredIndices) == 0 {
		m.viewportStart = 0
		return m
	}
	visible := m.listHeight()
	// Cursor above viewport: scroll up
	if m.cursor < m.viewportStart {
		m.viewportStart = m.cursor
	}
	// Cursor below viewport: scroll down
	if m.cursor >= m.viewportStart+visible {
		m.viewportStart = m.cursor - visible + 1
	}
	// Clamp viewport start
	maxStart := max(len(m.filteredIndices)-visible, 0)
	if m.viewportStart < 0 {
		m.viewportStart = 0
	} else if m.viewportStart > maxStart {
		m.viewportStart = maxStart
	}
	return m
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
		m = m.adjustViewport()
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
		m = m.adjustViewport()
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
	b.WriteString("\n")

	dimStyle := lipgloss.NewStyle().Foreground(colorGray)
	chevronStyle := lipgloss.NewStyle().Foreground(colorGray)

	if len(m.entries) == 0 {
		b.WriteString("\n")
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
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  No entries match the filter."))
		b.WriteString("\n")
		return b.String()
	}

	// Show ▲ chevron if there are entries above the viewport
	if m.viewportStart > 0 {
		b.WriteString(chevronStyle.Render("▲"))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")
	}

	// Calculate visible range
	visibleHeight := m.listHeight()
	viewportEnd := m.viewportStart + visibleHeight
	if viewportEnd > len(entriesToShow) {
		viewportEnd = len(entriesToShow)
	}

	// Render only visible entries
	for displayIdx := m.viewportStart; displayIdx < viewportEnd; displayIdx++ {
		actualIdx := entriesToShow[displayIdx]
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

		var name string
		var label string

		matchedIndexes := m.fuzzyMatches[actualIdx]
		if len(matchedIndexes) > 0 {
			// Highlight matches using precomputed positions
			name = highlightMatches(matchedIndexes, entry.Name, 0, nameStyle)
			// Highlight matches in label
			if entry.Label != "" {
				labelOffset := len(entry.Name) + 1
				label = " " + highlightMatches(
					matchedIndexes,
					entry.Label,
					labelOffset,
					labelStyle,
				)
			}
		} else {
			name = nameStyle.Render(entry.Name)
			if entry.Label != "" {
				label = " " + labelStyle.Render(entry.Label)
			}
		}

		fmt.Fprintf(&b, "%s%s%s%s%s\n", cursor, groupPrefix, checkbox, name, label)
	}

	return b.String()
}

// hasEntriesBelow returns true if there are entries below the viewport.
func (m Model) hasEntriesBelow() bool {
	if len(m.filteredIndices) == 0 {
		return false
	}
	return m.viewportStart+m.listHeight() < len(m.filteredIndices)
}
