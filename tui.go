package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	colorBlack         lipgloss.ANSIColor = 0
	colorRed           lipgloss.ANSIColor = 1
	colorGreen         lipgloss.ANSIColor = 2
	colorYellow        lipgloss.ANSIColor = 3
	colorBlue          lipgloss.ANSIColor = 4
	colorMagenta       lipgloss.ANSIColor = 5
	colorCyan          lipgloss.ANSIColor = 6
	colorWhite         lipgloss.ANSIColor = 7
	colorGray          lipgloss.ANSIColor = 8
	colorBrightBlack   lipgloss.ANSIColor = 8
	colorBrightRed     lipgloss.ANSIColor = 9
	colorBrightGreen   lipgloss.ANSIColor = 10
	colorBrightYellow  lipgloss.ANSIColor = 11
	colorBrightBlue    lipgloss.ANSIColor = 12
	colorBrightMagenta lipgloss.ANSIColor = 13
	colorBrightCyan    lipgloss.ANSIColor = 14
	colorBrightWhite   lipgloss.ANSIColor = 15
)

// viewMode represents the current mode of the TUI.
type viewMode int

const (
	modeList viewMode = iota
	modeAdd
	modeEdit
	modeConfirmDelete
)

// inputField identifies which field is being edited in add/edit mode.
type inputField int

const (
	fieldName inputField = iota
	fieldValue
	fieldLabel
)

// Model is the bubbletea model for the apiki TUI.
type Model struct {
	entries      []Entry
	cursor       int
	mode         viewMode
	currentField inputField

	// Text inputs for add/edit mode
	nameInput  textinput.Model
	valueInput textinput.Model
	labelInput textinput.Model

	// editIndex tracks which entry is being edited (-1 for new)
	editIndex int

	// quitting indicates the user pressed 'q' to quit and apply
	quitting bool

	// cancelled indicates the user pressed Ctrl-C to abort
	cancelled bool

	// Terminal dimensions
	width  int
	height int

	// nameGroupsMemo memoizes the name groups for faster lookup
	nameGroupsMemo map[string][]int

	// Filter state
	filtering       bool
	filterInput     textinput.Model
	filteredIndices []int
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
func (m *Model) recomputeFilter() {
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
		return
	}

	// Try to find the target entry in the new filtered list
	for displayIdx, actualIdx := range m.filteredIndices {
		if actualIdx == targetEntryIdx {
			m.cursor = displayIdx
			return
		}
	}

	// Target entry was filtered out. Backtrack through entries to find one
	// that's still visible.
	for entryIdx := targetEntryIdx - 1; entryIdx >= 0; entryIdx-- {
		for displayIdx, actualIdx := range m.filteredIndices {
			if actualIdx == entryIdx {
				m.cursor = displayIdx
				return
			}
		}
	}

	// No earlier entry found, default to first entry
	m.cursor = 0
}

// NewModel creates a new Model with the given entries.
func NewModel(entries []Entry) Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "VAR_NAME"
	nameInput.CharLimit = 256

	valueInput := textinput.New()
	valueInput.Placeholder = "value"
	valueInput.CharLimit = 4096

	labelInput := textinput.New()
	labelInput.Placeholder = "description"
	labelInput.CharLimit = 256

	filterInput := textinput.New()
	filterInput.Placeholder = "Filter..."
	filterInput.CharLimit = 256

	SortEntries(entries)

	model := Model{
		entries:         entries,
		cursor:          0,
		mode:            modeList,
		nameInput:       nameInput,
		valueInput:      valueInput,
		labelInput:      labelInput,
		filterInput:     filterInput,
		editIndex:       -1,
		filteredIndices: make([]int, len(entries)),
	}
	model.recomputeFilter()
	model = model.updateInputWidths()
	return model
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.nameGroupsMemo = nil

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.updateInputWidths()
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeList:
			return m.updateList(msg)
		case modeAdd, modeEdit:
			return m.updateForm(msg)
		case modeConfirmDelete:
			return m.updateConfirmDelete(msg)
		}
	}

	return m, nil
}

// updateInputWidths sets the width of all text inputs based on available
// terminal width.
func (m Model) updateInputWidths() Model {
	for _, input := range []*textinput.Model{
		&m.nameInput,
		&m.valueInput,
		&m.labelInput,
		&m.filterInput,
	} {
		width := 2
		if input.Value() != "" {
			width = len(input.Value()) + 2
		} else if input.Placeholder != "" {
			width = len(input.Placeholder) + 2
		}
		input.Width = min(width, m.width-8)
	}
	return m
}

func (m *Model) clearFilter() {
	m.filtering = false
	m.filterInput.SetValue("")
	m.filterInput.Blur()
	m.recomputeFilter()
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
			m.clearFilter()
			return m, nil
		case "enter":
			m.filtering = false
			m.filterInput.Blur()
			m.recomputeFilter()
			return m, nil
		}
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.recomputeFilter()
		return m, cmd
	}

	// List mode keys
	switch key {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "q":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		// Clear filter if one is active
		if m.filterInput.Value() != "" {
			m.clearFilter()
		}
		return m, nil

	case "/":
		m.filtering = true
		m.filterInput.SetValue("")
		m.filterInput.Focus()
		m.recomputeFilter()
		return m, textinput.Blink

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.filteredIndices)-1 {
			m.cursor++
		}
		return m, nil

	case "enter":
		if len(m.filteredIndices) > 0 {
			m.clearFilter()
			actualIndex := m.filteredIndices[m.cursor]
			m.mode = modeEdit
			m.editIndex = actualIndex
			entry := m.entries[actualIndex]
			m.nameInput.SetValue(entry.Name)
			m.valueInput.SetValue(entry.Value)
			m.labelInput.SetValue(entry.Label)
			m.currentField = fieldName
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

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		m.clearFilter()
		m.mode = modeList
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
	case fieldValue:
		m.valueInput, cmd = m.valueInput.Update(msg)
	case fieldLabel:
		m.labelInput, cmd = m.labelInput.Update(msg)
	}

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
	entry := Entry{
		Name:     strings.TrimSpace(m.nameInput.Value()),
		Value:    m.valueInput.Value(),
		Label:    strings.TrimSpace(m.labelInput.Value()),
		Selected: false,
	}

	if entry.Name == "" {
		// Don't save entries without a name
		m.clearFilter()
		m.mode = modeList
		return m, nil
	}

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
	m.clearFilter()

	// Move cursor to the saved entry's new position
	for i, e := range m.entries {
		if e.Name == entry.Name && e.Label == entry.Label && e.Value == entry.Value {
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
	return m, nil
}

func (m Model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "y", "Y", "enter":
		if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
			actualIndex := m.filteredIndices[m.cursor]
			if actualIndex < len(m.entries) {
				m.entries = append(m.entries[:actualIndex], m.entries[actualIndex+1:]...)
				m.recomputeFilter()
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

// View implements tea.Model.
func (m Model) View() string {
	var b strings.Builder

	switch m.mode {
	case modeList:
		b.WriteString(m.viewList())
	case modeAdd:
		b.WriteString(m.viewForm("Add Entry"))
	case modeEdit:
		b.WriteString(m.viewForm("Edit Entry"))
	case modeConfirmDelete:
		b.WriteString(m.viewConfirmDelete())
	}

	b.WriteString("\n")
	if m.mode == modeList && (m.filtering || m.filterInput.Value() != "") {
		b.WriteString(m.viewFilterBar())
		b.WriteString("\n")
	}
	b.WriteString(m.viewHelpBar())

	return b.String()
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

func (m Model) viewForm(title string) string {
	var b strings.Builder

	titleStyle := lipgloss.
		NewStyle().
		Bold(true).
		Foreground(colorBrightBlue)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().Width(8)

	b.WriteString(labelStyle.Render("Name:"))
	b.WriteString(m.nameInput.View())
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Value:"))
	b.WriteString(m.valueInput.View())
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Label:"))
	b.WriteString(m.labelInput.View())
	b.WriteString("\n")

	return b.String()
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

func (m Model) viewHelpBar() string {
	keyStyle := lipgloss.NewStyle().
		Background(colorCyan).
		Foreground(colorBlack).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(colorWhite).
		MarginRight(2)

	var items []string

	switch m.mode {
	case modeList:
		if m.filtering {
			items = []string{
				keyStyle.Render("Enter") + labelStyle.Render("Apply"),
				keyStyle.Render("Esc") + labelStyle.Render("Cancel"),
			}
		} else {
			items = []string{
				keyStyle.Render("/") + labelStyle.Render("Filter"),
				keyStyle.Render("↑↓") + labelStyle.Render("Move"),
				keyStyle.Render("Space") + labelStyle.Render("Toggle"),
				keyStyle.Render("+") + labelStyle.Render("Add"),
				keyStyle.Render("Enter") + labelStyle.Render("Edit"),
				keyStyle.Render("-") + labelStyle.Render("Delete"),
				keyStyle.Render("q") + labelStyle.Render("Quit"),
			}
			if m.filterInput.Value() != "" {
				// Insert "Esc Clear" after "/" when filter is active
				items = []string{
					keyStyle.Render("/") + labelStyle.Render("Filter"),
					keyStyle.Render("Esc") + labelStyle.Render("Clear"),
					keyStyle.Render("↑↓") + labelStyle.Render("Move"),
					keyStyle.Render("Space") + labelStyle.Render("Toggle"),
					keyStyle.Render("+") + labelStyle.Render("Add"),
					keyStyle.Render("Enter") + labelStyle.Render("Edit"),
					keyStyle.Render("-") + labelStyle.Render("Delete"),
					keyStyle.Render("q") + labelStyle.Render("Quit"),
				}
			}
		}
	case modeAdd, modeEdit:
		items = []string{
			keyStyle.Render("Tab/↓") + labelStyle.Render("Next"),
			keyStyle.Render("Shift+Tab/↑") + labelStyle.Render("Prev"),
			keyStyle.Render("Enter") + labelStyle.Render("Save"),
			keyStyle.Render("Esc") + labelStyle.Render("Cancel"),
		}
	case modeConfirmDelete:
		items = []string{
			keyStyle.Render("y/Enter") + labelStyle.Render("Yes"),
			keyStyle.Render("n/Esc") + labelStyle.Render("No"),
		}
	}

	return strings.Join(items, "")
}

// Quitting returns true if the user quit with 'q' (apply changes).
func (m Model) Quitting() bool {
	return m.quitting
}

// Cancelled returns true if the user quit with Ctrl-C (discard changes).
func (m Model) Cancelled() bool {
	return m.cancelled
}

// Entries returns the current entries (potentially modified).
func (m Model) Entries() []Entry {
	return m.entries
}
