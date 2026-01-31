package config

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/loderunner/apiki/internal/entries"
	"github.com/loderunner/apiki/internal/set"
)

func TestMain(m *testing.M) {
	// Use an in-memory filesystem for testing
	fs = afero.NewMemMapFs()
	os.Exit(m.Run())
}

func TestLoad(t *testing.T) {
	path := "/test/config.json"

	// Create a config file
	cfg := &Config{
		Selected: set.New("VAR1", "VAR2[0]"),
	}
	err := Save(path, cfg)
	require.NoError(t, err, "Save failed")

	// Load it back
	loaded, err := Load(path)
	require.NoError(t, err, "Load failed")
	require.NotNil(t, loaded)

	assert.Contains(t, loaded.Selected, "VAR1")
	assert.Contains(t, loaded.Selected, "VAR2[0]")
}

func TestLoadMissing(t *testing.T) {
	path := "/nonexistent.json"

	cfg, err := Load(path)
	require.NoError(t, err, "Load failed")
	require.NotNil(t, cfg, "Load returned nil config")

	assert.NotNil(t, cfg.Selected, "expected non-nil Selected set")
	assert.Empty(t, cfg.Selected, "expected empty Selected set")
}

func TestSave(t *testing.T) {
	path := "/test/save-config.json"

	cfg := &Config{
		Selected: set.New("VAR1", "VAR2"),
	}

	err := Save(path, cfg)
	require.NoError(t, err, "Save failed")

	// Verify file exists
	exists, err := afero.Exists(fs, path)
	require.NoError(t, err)
	assert.True(t, exists, "config file not created")
}

func TestEntryID_UniqueName(t *testing.T) {
	entries := []entries.Entry{
		{Name: "VAR1", Value: "value1"},
		{Name: "VAR2", Value: "value2"},
		{Name: "VAR3", Value: "value3"},
	}

	assert.Equal(t, "VAR1", EntryID(entries, 0))
	assert.Equal(t, "VAR2", EntryID(entries, 1))
	assert.Equal(t, "VAR3", EntryID(entries, 2))
}

func TestEntryID_RadioGroup(t *testing.T) {
	entries := []entries.Entry{
		{Name: "VAR", Value: "value1"},
		{Name: "VAR", Value: "value2"},
		{Name: "VAR", Value: "value3"},
		{Name: "OTHER", Value: "other"},
		{Name: "VAR", Value: "value4"},
	}

	assert.Equal(t, "VAR[0]", EntryID(entries, 0))
	assert.Equal(t, "VAR[1]", EntryID(entries, 1))
	assert.Equal(t, "VAR[2]", EntryID(entries, 2))
	assert.Equal(t, "OTHER", EntryID(entries, 3))
	assert.Equal(t, "VAR[3]", EntryID(entries, 4))
}

func TestConfigRoundTrip(t *testing.T) {
	path := "/test/roundtrip-config.json"

	original := &Config{
		Selected: set.New("VAR1", "VAR2[0]", "VAR3"),
	}

	err := Save(path, original)
	require.NoError(t, err, "Save failed")

	loaded, err := Load(path)
	require.NoError(t, err, "Load failed")
	require.NotNil(t, loaded)

	assert.Len(
		t,
		loaded.Selected,
		len(original.Selected),
		"Selected length mismatch",
	)

	for _, member := range original.Selected.Members() {
		assert.Contains(
			t,
			loaded.Selected.Members(),
			member,
			"Selected missing member: %s",
			member,
		)
	}
}
