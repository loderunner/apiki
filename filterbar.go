package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// fuzzyMatch checks if query matches entry using sparse case-insensitive
// matching. The query characters must appear in order within "name label".
func fuzzyMatch(entry Entry, query string) bool {
	if query == "" {
		return true
	}

	target := entry.Name + " " + entry.Label
	targetLower := strings.ToLower(target)
	queryLower := strings.ToLower(query)

	queryIdx := 0
	for i := 0; i < len(targetLower) && queryIdx < len(queryLower); i++ {
		if targetLower[i] == queryLower[queryIdx] {
			queryIdx++
		}
	}

	return queryIdx == len(queryLower)
}

// recomputeFilter updates filteredIndices based on the current filter query.
// The cursor persists on the same entry when possible. If the current entry
// gets filtered out, backtrack through entries to find a visible one.
func (m Model) recomputeFilter() Model {
	// Remember the actual entry index the cursor is pointing to
	var targetEntryIdx int
	if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
		targetEntryIdx = m.filteredIndices[m.cursor]
	}

	// Recompute filtered indices
	query := strings.TrimSpace(m.filterInput.Value())
	if query == "" {
		m.filteredIndices = make([]int, len(m.entries))
		for i := range m.entries {
			m.filteredIndices[i] = i
		}
	} else {
		m.filteredIndices = m.filteredIndices[:0]
		for i, entry := range m.entries {
			if fuzzyMatch(entry, query) {
				m.filteredIndices = append(m.filteredIndices, i)
			}
		}
	}

	if len(m.filteredIndices) == 0 {
		m.cursor = 0
		return m
	}

	// Try to find the target entry in the new filtered list
	for displayIdx, actualIdx := range m.filteredIndices {
		if actualIdx == targetEntryIdx {
			m.cursor = displayIdx
			return m
		}
	}

	// Target entry was filtered out. Backtrack through entries to find one
	// that's still visible.
	for entryIdx := targetEntryIdx - 1; entryIdx >= 0; entryIdx-- {
		for displayIdx, actualIdx := range m.filteredIndices {
			if actualIdx == entryIdx {
				m.cursor = displayIdx
				return m
			}
		}
	}

	// No earlier entry found, default to first entry
	m.cursor = 0
	return m
}

func (m Model) viewFilterBar() string {
	var b strings.Builder

	filterStyle := lipgloss.NewStyle().Foreground(colorBrightCyan)
	countStyle := lipgloss.NewStyle().Foreground(colorGray)

	b.WriteString(filterStyle.Render("Filter: "))
	b.WriteString(m.filterInput.View())

	matchCount := len(m.filteredIndices)
	totalCount := len(m.entries)
	countText := fmt.Sprintf("(%d/%d entries)", matchCount, totalCount)
	b.WriteString(" ")
	b.WriteString(countStyle.Render(countText))

	return b.String()
}
