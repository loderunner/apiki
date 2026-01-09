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

	return Model{
		entries:    entries,
		cursor:     0,
		mode:       modeList,
		nameInput:  nameInput,
		valueInput: valueInput,
		labelInput: labelInput,
		editIndex:  -1,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit

	case "q":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}

	case " ":
		if len(m.entries) > 0 {
			m.entries[m.cursor].Selected = !m.entries[m.cursor].Selected
		}

	case "+":
		m.mode = modeAdd
		m.editIndex = -1
		m.nameInput.SetValue("")
		m.valueInput.SetValue("")
		m.labelInput.SetValue("")
		m.currentField = fieldName
		m.nameInput.Focus()
		return m, textinput.Blink

	case "enter":
		if len(m.entries) > 0 {
			m.mode = modeEdit
			m.editIndex = m.cursor
			entry := m.entries[m.cursor]
			m.nameInput.SetValue(entry.Name)
			m.valueInput.SetValue(entry.Value)
			m.labelInput.SetValue(entry.Label)
			m.currentField = fieldName
			m.nameInput.Focus()
			return m, textinput.Blink
		}

	case "backspace", "delete", "-":
		if len(m.entries) > 0 {
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
		m.cursor = len(m.entries) - 1
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
		if m.cursor < len(m.entries) {
			m.entries = append(m.entries[:m.cursor], m.entries[m.cursor+1:]...)
			if m.cursor >= len(m.entries) && m.cursor > 0 {
				m.cursor--
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

	if len(m.entries) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(colorGray)
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

	for i, entry := range m.entries {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		var checkbox string
		if entry.Selected {
			checkbox = selectedStyle.Render("⦿ ")
		} else {
			checkbox = unselectedStyle.Render("◯ ")
		}

		name := nameStyle.Render(entry.Name)
		label := ""
		if entry.Label != "" {
			label = " " + labelStyle.Render(entry.Label)
		}

		fmt.Fprintf(&b, "%s%s%s%s\n", cursor, checkbox, name, label)
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

	hintStyle := lipgloss.NewStyle().Foreground(colorGray)
	b.WriteString("\n")
	b.WriteString(
		hintStyle.Render(
			"Tab/↓ next field • Shift+Tab/↑ prev • Enter save • Esc cancel",
		),
	)

	return b.String()
}

func (m Model) viewConfirmDelete() string {
	var b strings.Builder

	if m.cursor < 0 || m.cursor >= len(m.entries) {
		return ""
	}

	entry := m.entries[m.cursor]

	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBrightYellow)
	b.WriteString(warnStyle.Render("Delete Entry?"))
	b.WriteString("\n\n")

	nameStyle := lipgloss.NewStyle().Bold(true)
	// fmt.Fprintf(&b, "  %s", nameStyle.Render(entry.Name))
	b.WriteString(fmt.Sprintf("  %s", nameStyle.Render(entry.Name)))
	if entry.Label != "" {
		labelStyle := lipgloss.NewStyle().Foreground(colorGray).Italic(true)
		fmt.Fprintf(&b, " %s", labelStyle.Render(entry.Label))
	}
	b.WriteString("\n\n")

	hintStyle := lipgloss.NewStyle().Foreground(colorGray)
	b.WriteString(hintStyle.Render("Press y to confirm, n or Esc to cancel"))

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
		items = []string{
			keyStyle.Render("Space") + labelStyle.Render("Toggle"),
			keyStyle.Render("+") + labelStyle.Render("Add"),
			keyStyle.Render("Enter") + labelStyle.Render("Edit"),
			keyStyle.Render("-") + labelStyle.Render("Delete"),
			keyStyle.Render("q") + labelStyle.Render("Quit"),
		}
	case modeAdd, modeEdit:
		items = []string{
			keyStyle.Render("Tab") + labelStyle.Render("Next"),
			keyStyle.Render("Enter") + labelStyle.Render("Save"),
			keyStyle.Render("Esc") + labelStyle.Render("Cancel"),
		}
	case modeConfirmDelete:
		items = []string{
			keyStyle.Render("y") + labelStyle.Render("Yes"),
			keyStyle.Render("n") + labelStyle.Render("No"),
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
