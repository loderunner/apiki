package restore

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/loderunner/apiki/commands/testutil"
	"github.com/loderunner/apiki/internal/config"
	"github.com/loderunner/apiki/internal/entries"
	"github.com/loderunner/apiki/internal/keychain"
	"github.com/loderunner/apiki/internal/prompt"
	"github.com/loderunner/apiki/internal/set"
)

func TestMain(m *testing.M) {
	entries.UseFs(afero.NewMemMapFs())
	config.UseFs(afero.NewMemMapFs())
	os.Exit(m.Run())
}

func TestRestorePlainFile(t *testing.T) {
	file := &entries.File{
		Entries: []entries.Entry{
			{Name: "FOO", Value: "bar"},
			{Name: "BAZ", Value: "qux"},
		},
	}
	err := entries.Save("/tmp/restore/vars.json", file)
	require.NoError(t, err)

	cfg := &config.Config{
		Selected: set.New("FOO"),
	}
	err = config.Save("/tmp/restore/config.json", cfg)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(nil, nil))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	output, err := Run(ctx, "/tmp/restore/vars.json", "/tmp/restore/config.json")
	require.NoError(t, err)
	require.Contains(t, output, "export FOO='bar'")
	require.NotContains(t, output, "BAZ")
}

func TestRestoreEncryptedFile(t *testing.T) {
	file := &entries.File{
		Entries: []entries.Entry{
			{Name: "SECRET", Value: "value"},
		},
	}
	key, err := file.SetPasswordMode("mypass")
	require.NoError(t, err)
	err = file.EncryptValues(key)
	require.NoError(t, err)
	err = entries.Save("/tmp/restore2/vars.json", file)
	require.NoError(t, err)

	cfg := &config.Config{
		Selected: set.New("SECRET"),
	}
	err = config.Save("/tmp/restore2/config.json", cfg)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter([]string{"mypass"}, nil))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	output, err := Run(ctx, "/tmp/restore2/vars.json", "/tmp/restore2/config.json")
	require.NoError(t, err)
	require.Contains(t, output, "export SECRET='value'")
}

func TestRestoreEmptySelection(t *testing.T) {
	file := &entries.File{
		Entries: []entries.Entry{{Name: "FOO", Value: "bar"}},
	}
	err := entries.Save("/tmp/restore3/vars.json", file)
	require.NoError(t, err)

	cfg := &config.Config{
		Selected: set.New[string](),
	}
	err = config.Save("/tmp/restore3/config.json", cfg)
	require.NoError(t, err)

	ctx := context.Background()
	output, err := Run(ctx, "/tmp/restore3/vars.json", "/tmp/restore3/config.json")
	require.NoError(t, err)
	require.Empty(t, output)
}
