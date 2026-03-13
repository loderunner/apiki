package apiki

import (
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"

	"github.com/loderunner/apiki/internal/entries"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

func TestTUISelectAndApply(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: false},
		{Entry: entries.Entry{Name: "BAZ", Value: "qux"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		_ = tm.Quit()
	})

	// Select first entry
	tm.Send(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	// Apply and quit
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(2*time.Second))
	requireModel, ok := fm.(Model)
	require.True(t, ok, "expected Model type")
	require.True(t, requireModel.Quitting())
	require.True(t, requireModel.Entries()[0].Selected)
}

func TestTUIAddEntry(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "A", Value: "1"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		_ = tm.Quit()
	})

	// Add new entry
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	tm.Type("NEWVAR")
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Type("newvalue")
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Type("label")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Cancel to exit (we're back in list mode)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(2*time.Second))
	requireModel, ok := fm.(Model)
	require.True(t, ok)
	require.Len(t, requireModel.Entries(), 2)
	var found bool
	for _, e := range requireModel.Entries() {
		if e.Name == "NEWVAR" && e.Value == "newvalue" {
			found = true
			break
		}
	}
	require.True(t, found)

	file, err := entries.Load("/tmp/test/variables.json")
	require.NoError(t, err)
	require.Len(t, file.Entries, 2)
}

func TestTUIDeleteEntry(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "A", Value: "1"}, Selected: false},
		{Entry: entries.Entry{Name: "B", Value: "2"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		_ = tm.Quit()
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(2*time.Second))
	requireModel, ok := fm.(Model)
	require.True(t, ok)
	require.Len(t, requireModel.Entries(), 1)
	require.Equal(t, "B", requireModel.Entries()[0].Name)
}

func TestTUIFilterAndSelect(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "1"}, Selected: false},
		{Entry: entries.Entry{Name: "BAR", Value: "2"}, Selected: false},
		{Entry: entries.Entry{Name: "BAZ", Value: "3"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		_ = tm.Quit()
	})

	// Enter filter mode, type "ba", Enter to exit filter input (list stays
	// filtered)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	tm.Type("ba")
	tm.Send(
		tea.KeyMsg{Type: tea.KeyEnter},
	) // exit filter input - now we can use Space
	tm.Send(
		tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")},
	) // select first filtered (BAR)
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape}) // clear filter
	tm.Send(
		tea.KeyMsg{Type: tea.KeyEnter},
	) // apply and quit

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(2*time.Second))
	requireModel, ok := fm.(Model)
	require.True(t, ok)
	require.True(t, requireModel.Quitting())
	// After filtering "ba" and selecting, one of BAR or BAZ should be selected
	selectedCount := 0
	for _, e := range requireModel.Entries() {
		if e.Selected {
			selectedCount++
			require.Contains(t, []string{"BAR", "BAZ"}, e.Name)
		}
	}
	require.Equal(t, 1, selectedCount, "expected exactly one selected entry")
}

func TestTUICtrlCCancelled(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		_ = tm.Quit()
	})

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(2*time.Second))
	requireModel, ok := fm.(Model)
	require.True(t, ok)
	require.True(t, requireModel.Cancelled())
}

func TestTUIOutputContainsExpectedElements(t *testing.T) {
	testEntries := []Entry{
		{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: false},
	}
	m := newTestModel(t, testEntries)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		_ = tm.Quit()
	})

	// Give it a moment to render
	time.Sleep(50 * time.Millisecond)

	out := readBts(t, tm.Output())
	require.Contains(t, string(out), "Environment Variables")
	require.Contains(t, string(out), "FOO")
	require.Contains(t, string(out), "Filter")
}

func readBts(tb testing.TB, r io.Reader) []byte {
	tb.Helper()
	bts, err := io.ReadAll(r)
	if err != nil {
		tb.Fatal(err)
	}
	return bts
}
