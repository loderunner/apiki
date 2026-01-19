package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// entrySource implements fuzzy.Source for []Entry.
type entrySource []Entry

func (s entrySource) Len() int {
	return len(s)
}

func (s entrySource) String(i int) string {
	return s[i].FuzzyTarget()
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
		m.fuzzyMatches = nil
	} else {
		matches := fuzzy.FindFrom(query, entrySource(m.entries))
		m.filteredIndices = make([]int, len(matches))
		m.fuzzyMatches = make(map[int][]int)
		for i, match := range matches {
			m.filteredIndices[i] = match.Index
			m.fuzzyMatches[match.Index] = match.MatchedIndexes
		}
	}

	if len(m.filteredIndices) == 0 {
		m.cursor = 0
		m = m.adjustViewport()
		return m
	}

	// Try to find the target entry in the new filtered list
	for displayIdx, actualIdx := range m.filteredIndices {
		if actualIdx == targetEntryIdx {
			m.cursor = displayIdx
			m = m.adjustViewport()
			return m
		}
	}

	// Target entry was filtered out. Backtrack through entries to find one
	// that's still visible.
	for entryIdx := targetEntryIdx - 1; entryIdx >= 0; entryIdx-- {
		for displayIdx, actualIdx := range m.filteredIndices {
			if actualIdx == entryIdx {
				m.cursor = displayIdx
				m = m.adjustViewport()
				return m
			}
		}
	}

	// No earlier entry found, default to first entry
	m.cursor = 0
	m = m.adjustViewport()
	return m
}

func (m Model) viewFilterBar() string {
	var b strings.Builder

	filterStyle := lipgloss.NewStyle().Foreground(ColorBrightCyan)
	countStyle := lipgloss.NewStyle().Foreground(ColorGray)

	b.WriteString(filterStyle.Render("Filter: "))
	b.WriteString(m.filterInput.View())

	matchCount := len(m.filteredIndices)
	totalCount := len(m.entries)
	countText := fmt.Sprintf("(%d/%d entries)", matchCount, totalCount)
	b.WriteString(" ")
	b.WriteString(countStyle.Render(countText))

	return b.String()
}
