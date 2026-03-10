package apiki

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/loderunner/apiki/internal/config"
	"github.com/loderunner/apiki/internal/entries"
)

func TestMain(m *testing.M) {
	entries.UseFs(afero.NewMemMapFs())
	config.UseFs(afero.NewMemMapFs())
	os.Exit(m.Run())
}

func newTestModel(t *testing.T, testEntries []Entry) Model {
	t.Helper()
	file := &entries.File{
		Entries: make([]entries.Entry, 0),
	}
	for _, e := range testEntries {
		if e.SourceFile == "" {
			file.Entries = append(file.Entries, entries.Entry{
				Name:  e.Name,
				Value: e.Value,
				Label: e.Label,
			})
		}
	}
	return NewModel(file, "/tmp/test/variables.json", "/tmp/test/config.json", nil, testEntries)
}

func sendKey(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	model, cmd := m.Update(key)
	return model.(Model), cmd
}

func sendWindowSize(m Model, w, h int) (Model, tea.Cmd) {
	model, cmd := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return model.(Model), cmd
}

func runUpdate(m Model, msg tea.Msg) Model {
	model, _ := m.Update(msg)
	return model.(Model)
}

func TestModel_ListNavigation(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "A", Value: "1"}, Selected: false},
		{Entry: entries.Entry{Name: "B", Value: "2"}, Selected: false},
		{Entry: entries.Entry{Name: "C", Value: "3"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	// Down moves cursor
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, 1, m.cursor)

	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, 2, m.cursor)

	// Down at bottom wraps to top
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, 0, m.cursor)

	// Up at top wraps to bottom
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, 2, m.cursor)
}

func TestModel_Selection(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: false},
		{Entry: entries.Entry{Name: "BAZ", Value: "qux"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	// Space toggles selection
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	require.True(t, m.entries[0].Selected)

	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	require.False(t, m.entries[0].Selected)
}

func TestModel_SelectionRadioGroup(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "ENV", Value: "dev"}, Selected: false},
		{Entry: entries.Entry{Name: "ENV", Value: "prod"}, Selected: false},
		{Entry: entries.Entry{Name: "ENV", Value: "staging"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	// Select first
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	require.True(t, m.entries[0].Selected)
	require.False(t, m.entries[1].Selected)
	require.False(t, m.entries[2].Selected)

	// Select second - deselects first
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	require.False(t, m.entries[0].Selected)
	require.True(t, m.entries[1].Selected)
	require.False(t, m.entries[2].Selected)
}

func TestModel_AddEntry(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "A", Value: "1"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	// Enter add mode
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	require.Equal(t, modeAdd, m.mode)

	// Esc cancels
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyEscape})
	require.Equal(t, modeList, m.mode)
	require.Len(t, m.entries, 1)

	// Add new entry
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	// Type name
	for _, r := range "NEWVAR" {
		m = runUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyTab})
	// Type value
	for _, r := range "newvalue" {
		m = runUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyTab})
	// Type label (optional)
	for _, r := range "new entry" {
		m = runUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyEnter})

	require.Equal(t, modeList, m.mode)
	require.Len(t, m.entries, 2)
	var found bool
	for _, e := range m.entries {
		if e.Name == "NEWVAR" && e.Value == "newvalue" {
			found = true
			break
		}
	}
	require.True(t, found, "new entry should be in list")

	// Verify file was persisted
	file, err := entries.Load("/tmp/test/variables.json")
	require.NoError(t, err)
	require.Len(t, file.Entries, 2)
}

func TestModel_EditEntry(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "bar", Label: "old"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	// Enter edit mode
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("=")})
	require.Equal(t, modeEdit, m.mode)
	require.Equal(t, "FOO", m.nameInput.Value())
	require.Equal(t, "bar", m.valueInput.Value())

	// Tab to value field, clear and type new value, Tab to label, Enter to save
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyTab})
	for range "bar" {
		m = runUpdate(m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	for _, r := range "updated" {
		m = runUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyTab})
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyEnter})

	require.Equal(t, modeList, m.mode)
	require.Equal(t, "updated", m.entries[0].Value)
}

func TestModel_DeleteEntry(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "A", Value: "1"}, Selected: false},
		{Entry: entries.Entry{Name: "B", Value: "2"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	// Enter delete confirmation
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	require.Equal(t, modeConfirmDelete, m.mode)

	// n cancels
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	require.Equal(t, modeList, m.mode)
	require.Len(t, m.entries, 2)

	// Delete for real
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	require.Equal(t, modeConfirmDelete, m.mode)
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	require.Equal(t, modeList, m.mode)
	require.Len(t, m.entries, 1)
	require.Equal(t, "B", m.entries[0].Name)
}

func TestModel_Filter(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "1"}, Selected: false},
		{Entry: entries.Entry{Name: "BAR", Value: "2"}, Selected: false},
		{Entry: entries.Entry{Name: "BAZ", Value: "3"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	// Enter filter mode
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	require.True(t, m.filtering)

	// Type filter
	for _, r := range "ba" {
		m = runUpdate(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	require.Len(t, m.filteredIndices, 2) // BAR and BAZ

	// Esc clears filter
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyEscape})
	require.False(t, m.filtering)
	require.Len(t, m.filteredIndices, 3)
}

func TestModel_QuitAndApply(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	// Select and quit
	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	m, cmd := sendKey(m, tea.KeyMsg{Type: tea.KeyEnter})

	require.True(t, m.Quitting())
	require.NotNil(t, cmd)

	// Config should be persisted (selection)
	cfg, err := config.Load("/tmp/test/config.json")
	require.NoError(t, err)
	require.True(t, cfg.Selected.Has("FOO"))
}

func TestModel_Cancel(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	m, _ = sendKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	require.True(t, m.Cancelled())
}

func TestModel_ViewOutput(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	m, _ = sendWindowSize(m, 80, 24)

	view := m.View()
	require.Contains(t, view, "Environment Variables")
	require.Contains(t, view, "FOO")
	require.Contains(t, view, "Filter")
	require.Contains(t, view, "Apply")
}
