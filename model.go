package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

	// Validation errors for form fields
	nameError  string
	valueError string

	// nameGroupsMemo memoizes the name groups for faster lookup
	nameGroupsMemo map[string][]int

	// Filter state
	filtering       bool
	filterInput     textinput.Model
	filteredIndices []int
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
	model = model.recomputeFilter()
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

func (m Model) clearFilter() Model {
	m.filtering = false
	m.filterInput.SetValue("")
	m.filterInput.Blur()
	return m.recomputeFilter()
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
