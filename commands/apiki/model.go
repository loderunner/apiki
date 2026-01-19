package apiki

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/loderunner/apiki/internal/entries"
)

// viewMode represents the current mode of the TUI.
type viewMode int

const (
	modeList viewMode = iota
	modeAdd
	modeEdit
	modeConfirmDelete
	modeConfirmPromote
	modeConfirmImport
	modeError
	modeImport
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
	// file holds the apiki entries file (decrypted in memory)
	file *entries.File

	// filePath is the path to the entries file
	filePath string

	// encryptionKey is the encryption key (nil if unencrypted)
	encryptionKey []byte

	// entries holds all entries (apiki + .env) for TUI display
	entries []Entry

	cursor int
	mode   viewMode

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

	// errorMessage stores an error message to display in error mode
	errorMessage string

	// nameGroupsMemo memoizes the name groups for faster lookup
	nameGroupsMemo map[string][]int

	// Filter state
	filtering       bool
	filterInput     textinput.Model
	filteredIndices []int
	fuzzyMatches    map[int][]int

	// Viewport state for scrolling list
	viewportStart int // first visible entry index in list mode

	// Import mode state
	originalEntries []Entry // stored entries when in import mode
}

// NewModel creates a new Model with the given file, file path, encryption key,
// and combined entries (apiki + .env) for TUI display.
func NewModel(
	file *entries.File,
	filePath string,
	encryptionKey []byte,
	allEntries []Entry,
) Model {
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

	SortEntries(allEntries)

	model := Model{
		file:            file,
		filePath:        filePath,
		encryptionKey:   encryptionKey,
		entries:         allEntries,
		cursor:          0,
		mode:            modeList,
		nameInput:       nameInput,
		valueInput:      valueInput,
		labelInput:      labelInput,
		filterInput:     filterInput,
		editIndex:       -1,
		filteredIndices: make([]int, len(allEntries)),
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
		if m.mode == modeList || m.mode == modeImport {
			m = m.adjustViewport()
		}
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeList, modeImport:
			return m.updateList(msg)
		case modeAdd, modeEdit:
			return m.updateForm(msg)
		case modeConfirmDelete:
			return m.updateConfirmDelete(msg)
		case modeConfirmPromote:
			return m.updateConfirmPromote(msg)
		case modeConfirmImport:
			return m.updateConfirmImport(msg)
		case modeError:
			return m.updateError(msg)
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	var b strings.Builder

	switch m.mode {
	case modeList, modeImport:
		b.WriteString(m.viewList())
	case modeAdd:
		b.WriteString(m.viewForm("Add Variable"))
	case modeEdit:
		b.WriteString(m.viewForm("Edit Variable"))
	case modeConfirmDelete:
		b.WriteString(m.viewConfirmDelete())
	case modeConfirmPromote:
		b.WriteString(m.viewConfirmPromote())
	case modeConfirmImport:
		b.WriteString(m.viewConfirmImport())
	case modeError:
		b.WriteString(m.viewError())
	}

	// Render bottom line: may contain ▼ chevron and/or filter bar
	if m.mode == modeList || m.mode == modeImport {
		hasFilter := m.filtering || m.filterInput.Value() != ""
		hasMore := m.hasEntriesBelow()

		if hasMore || hasFilter {
			chevronStyle := lipgloss.NewStyle().Foreground(ColorGray)
			if hasMore {
				b.WriteString(chevronStyle.Render("▼"))
				if hasFilter {
					b.WriteString(" ")
				}
			}
			if hasFilter {
				b.WriteString(m.viewFilterBar())
			}
			b.WriteString("\n")
		} else {
			b.WriteString("\n")
		}
	} else {
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

// prepareForm sets up the form for editing or adding an entry.
// If entry is provided, its values are pre-filled; otherwise fields are
// cleared. Returns the updated model and a command to start text input
// blinking.
func (m Model) prepareForm(editIndex int, entry *Entry) (Model, tea.Cmd) {
	m.editIndex = editIndex
	m.currentField = fieldName
	m.nameError = ""
	m.valueError = ""

	if entry != nil {
		m.nameInput.SetValue(entry.Name)
		m.valueInput.SetValue(entry.Value)
		m.labelInput.SetValue(entry.Label)
	} else {
		m.nameInput.SetValue("")
		m.valueInput.SetValue("")
		m.labelInput.SetValue("")
	}

	m = m.updateInputWidths()
	m.nameInput.Focus()
	return m, textinput.Blink
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

// persistEntries saves the current entries to the configured file path.
// Only saves apiki entries (those without SourceFile).
// Re-encrypts values if encryption is enabled.
// On error, switches to error mode to display the message.
func (m Model) persistEntries() Model {
	// Work on a copy to avoid mutating the in-memory state
	toSave := m.file.Clone()

	// Extract apiki entries from combined entries (those without SourceFile)
	apikiEntries := make([]entries.Entry, 0)
	for _, entry := range m.entries {
		if entry.SourceFile == "" {
			apikiEntries = append(apikiEntries, entries.Entry{
				Name:  entry.Name,
				Value: entry.Value,
				Label: entry.Label,
			})
		}
	}

	// Update file entries
	toSave.Entries = apikiEntries

	// Re-encrypt if encryption is enabled
	if toSave.Encrypted() && m.encryptionKey != nil {
		if err := toSave.EncryptValues(m.encryptionKey); err != nil {
			m.errorMessage = "Failed to encrypt variables: " + err.Error()
			m.mode = modeError
			return m
		}
	}

	// Save file
	if err := entries.Save(m.filePath, toSave); err != nil {
		m.errorMessage = "Failed to save variables: " + err.Error()
		m.mode = modeError
		return m
	}

	// Update in-memory file to match saved state (but keep decrypted)
	m.file.Entries = apikiEntries
	return m
}
