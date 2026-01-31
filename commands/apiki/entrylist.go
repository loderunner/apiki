package apiki

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/loderunner/apiki/internal/entries"
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
		Foreground(ColorBrightYellow).
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
	// The bottom line contains ▼ chevron and/or filter bar (they share the
	// line)
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
	case "q", "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		if m.filterInput.Value() != "" {
			m = m.clearFilter()
			return m, nil
		}
		if m.mode == modeImport {
			// Cancel import: restore original entries
			m.entries = m.originalEntries
			m.originalEntries = nil
			m.mode = modeList
			m = m.recomputeFilter()
			m = m.adjustViewport()
			return m, nil
		}
		m.cancelled = true
		return m, tea.Quit

	case "enter":
		if m.mode == modeImport {
			// Count selected entries
			selectedCount := 0
			for _, entry := range m.entries {
				if entry.Selected {
					selectedCount++
				}
			}
			// Only show confirmation if there are selected entries
			if selectedCount > 0 {
				m.mode = modeConfirmImport
			}
			return m, nil
		}
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
		// Don't allow editing in import mode
		if m.mode == modeImport {
			return m, nil
		}
		if len(m.filteredIndices) > 0 {
			m = m.clearFilter()
			actualIndex := m.filteredIndices[m.cursor]
			entry := m.entries[actualIndex]
			if entry.SourceFile != "" {
				// .env entry: show promote confirmation
				m.mode = modeConfirmPromote
			} else {
				// apiki entry: edit directly
				m.mode = modeEdit
				return m.prepareForm(actualIndex, &entry)
			}
		}
		return m, nil

	case " ":
		if len(m.filteredIndices) > 0 {
			actualIndex := m.filteredIndices[m.cursor]
			currentEntry := &m.entries[actualIndex]
			currentEntry.Selected = !currentEntry.Selected

			// Radio-button behavior: if we selected this entry, deselect others
			// with the same name (only in list mode, not import mode)
			if currentEntry.Selected && m.mode != modeImport {
				groups := m.nameGroups()
				for _, i := range groups[currentEntry.Name] {
					if i != actualIndex {
						m.entries[i].Selected = false
					}
				}
			}
		}

	case "+":
		// Don't allow creating new entries in import mode
		if m.mode == modeImport {
			return m, nil
		}
		m = m.clearFilter()
		m.mode = modeAdd
		return m.prepareForm(-1, nil)

	case "backspace", "delete", "-":
		// Don't allow deletion in import mode
		if m.mode == modeImport {
			return m, nil
		}
		if len(m.filteredIndices) > 0 {
			actualIndex := m.filteredIndices[m.cursor]
			entry := m.entries[actualIndex]
			// Block deletion of .env entries
			if entry.SourceFile == "" {
				m.mode = modeConfirmDelete
			}
			// Otherwise ignore the keypress for .env entries
		}

	case "i":
		if m.mode == modeList {
			// Store current entries
			m.originalEntries = make([]Entry, len(m.entries))
			copy(m.originalEntries, m.entries)

			// Load environment variables
			envEntries := loadEnvironmentEntries()
			m.entries = envEntries
			m.mode = modeImport
			m.cursor = 0
			m = m.clearFilter()
			m = m.recomputeFilter()
			m = m.adjustViewport()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) viewList() string {
	var b strings.Builder

	titleStyle := lipgloss.
		NewStyle().
		Bold(true).
		Foreground(ColorBrightBlue)
	title := "Environment Variables"
	if m.mode == modeImport {
		title = "Import from Environment"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	dimStyle := lipgloss.NewStyle().Foreground(ColorGray)
	chevronStyle := lipgloss.NewStyle().Foreground(ColorGray)

	if len(m.entries) == 0 {
		b.WriteString("\n")
		if m.mode == modeImport {
			b.WriteString(
				dimStyle.Render("  No environment variables to import."),
			)
		} else {
			b.WriteString(dimStyle.Render("  No entries. Press + to add one."))
		}
		b.WriteString("\n")
		return b.String()
	}

	selectedStyle := lipgloss.
		NewStyle().
		Foreground(ColorBrightGreen)
	unselectedStyle := lipgloss.
		NewStyle().
		Foreground(ColorGray)
	cursorStyle := lipgloss.NewStyle().Bold(true)
	nameStyle := lipgloss.NewStyle().Bold(true)
	labelStyle := lipgloss.
		NewStyle().
		Foreground(ColorGray).Italic(true)
	groupConnectorStyle := lipgloss.
		NewStyle().
		Foreground(ColorGray)

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

		// Check if this entry is part of a group (multiple entries with same
		// name)
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
				// For .env entries, the fuzzy target doesn't include "from "
				// prefix but the label does, so highlight only the part after
				// "from " to match the fuzzy target structure
				labelText := entry.Label
				labelPrefix := ""
				if entry.SourceFile != "" &&
					strings.HasPrefix(entry.Label, "from ") {
					labelPrefix = entry.Label[:5] // "from "
					labelText = entry.Label[5:]   // "dirname/filename"
				}
				highlightedLabel := highlightMatches(
					matchedIndexes,
					labelText,
					labelOffset,
					labelStyle,
				)
				label = " " + labelPrefix + highlightedLabel
			}
		} else {
			name = nameStyle.Render(entry.Name)
			if entry.Label != "" {
				label = " " + labelStyle.Render(entry.Label)
			}
		}

		fmt.Fprintf(
			&b,
			"%s%s%s%s%s\n",
			cursor,
			groupPrefix,
			checkbox,
			name,
			label,
		)
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

// loadEnvironmentEntries loads environment variables from os.Environ() and
// converts them to Entry format.
func loadEnvironmentEntries() []Entry {
	envVars := os.Environ()
	result := make([]Entry, 0, len(envVars))

	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		value := parts[1]

		result = append(result, Entry{
			Entry: entries.Entry{
				Name:  name,
				Value: value,
				Label: value,
			},
			Selected: false,
		})
	}

	SortEntries(result)
	return result
}

// confirmImport creates apiki entries for all selected environment variables
// and restores the original entries list.
func (m Model) confirmImport() (Model, tea.Cmd) {
	// Collect selected entries
	selectedEntries := make([]Entry, 0)
	for _, entry := range m.entries {
		if entry.Selected {
			// Create new apiki entry (no SourceFile)
			selectedEntries = append(selectedEntries, Entry{
				Entry: entries.Entry{
					Name:  entry.Name,
					Value: entry.Value,
					Label: "imported from environment",
				},
				Selected: true,
			})
		}
	}

	// Restore original entries
	m.entries = m.originalEntries
	m.originalEntries = nil

	// Add selected entries to the main list
	if len(selectedEntries) > 0 {
		m.entries = append(m.entries, selectedEntries...)
		SortEntries(m.entries)
		m = m.persistEntries()

		// On persist failure, stay in error mode
		if m.mode == modeError {
			return m, nil
		}
	}

	// Return to list mode
	m.mode = modeList
	m = m.recomputeFilter()
	m.cursor = 0
	m = m.adjustViewport()

	return m, nil
}
