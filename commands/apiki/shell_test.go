package apiki

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/loderunner/apiki/internal/entries"
)

func TestGenerateShellCommands(t *testing.T) {
	t.Run("export selected with new value", func(t *testing.T) {
		es := []Entry{
			{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: true},
			{Entry: entries.Entry{Name: "BAZ", Value: "qux"}, Selected: false},
		}
		envSnapshot := map[string]string{"FOO": "old", "BAZ": "original"}

		output := generateShellCommands(es, envSnapshot)
		require.Contains(t, output, "export FOO='bar'")
		require.Contains(t, output, "unset BAZ")
	})

	t.Run("no export when value unchanged", func(t *testing.T) {
		es := []Entry{
			{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: true},
		}
		envSnapshot := map[string]string{"FOO": "bar"}

		output := generateShellCommands(es, envSnapshot)
		require.NotContains(t, output, "export FOO")
		require.Empty(t, output)
	})

	t.Run("single quote escaping", func(t *testing.T) {
		es := []Entry{
			{
				Entry:    entries.Entry{Name: "FOO", Value: "it's working"},
				Selected: true,
			},
		}
		envSnapshot := map[string]string{}

		output := generateShellCommands(es, envSnapshot)
		require.Contains(t, output, "export FOO='it'\\''s working'")
	})

	t.Run("empty selection", func(t *testing.T) {
		es := []Entry{
			{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: false},
		}
		envSnapshot := map[string]string{"FOO": "old"}

		output := generateShellCommands(es, envSnapshot)
		require.Contains(t, output, "unset FOO")
	})

	t.Run("radio group selects one", func(t *testing.T) {
		es := []Entry{
			{Entry: entries.Entry{Name: "ENV", Value: "dev"}, Selected: false},
			{Entry: entries.Entry{Name: "ENV", Value: "prod"}, Selected: true},
			{
				Entry:    entries.Entry{Name: "ENV", Value: "staging"},
				Selected: false,
			},
		}
		envSnapshot := map[string]string{"ENV": "dev"}

		output := generateShellCommands(es, envSnapshot)
		require.Contains(t, output, "export ENV='prod'")
		require.NotContains(t, output, "unset ENV")
	})

	t.Run("no export for unset variable when deselected", func(t *testing.T) {
		es := []Entry{
			{Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: false},
		}
		envSnapshot := map[string]string{} // FOO was not set

		output := generateShellCommands(es, envSnapshot)
		require.Empty(t, output)
	})
}
