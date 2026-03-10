package encrypt

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/loderunner/apiki/commands/testutil"
	"github.com/loderunner/apiki/internal/entries"
	"github.com/loderunner/apiki/internal/keychain"
	"github.com/loderunner/apiki/internal/prompt"
)

func TestMain(m *testing.M) {
	entries.UseFs(afero.NewMemMapFs())
	os.Exit(m.Run())
}

func seedVariablesFile(t *testing.T, path string, file *entries.File) {
	t.Helper()
	err := entries.Save(path, file)
	require.NoError(t, err)
}

func TestEncryptPasswordMode(t *testing.T) {
	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(
		[]string{"secret", "secret"},
		[]string{"password"},
	))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	file := &entries.File{
		Entries: []entries.Entry{
			{Name: "FOO", Value: "bar"},
			{Name: "BAZ", Value: "qux"},
		},
	}
	seedVariablesFile(t, "/tmp/vars.json", file)

	err := Run(ctx, "/tmp/vars.json")
	require.NoError(t, err)

	loaded, err := entries.Load("/tmp/vars.json")
	require.NoError(t, err)
	require.True(t, loaded.Encrypted())
	require.Equal(t, "password", loaded.Encryption.Mode)
}

func TestEncryptKeychainMode(t *testing.T) {
	mockKC := &testutil.MockKeychain{}
	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(nil, []string{"keychain"}))
	ctx = keychain.WithKeychain(ctx, mockKC)

	file := &entries.File{
		Entries: []entries.Entry{
			{Name: "FOO", Value: "bar"},
		},
	}
	seedVariablesFile(t, "/tmp/vars2.json", file)

	err := Run(ctx, "/tmp/vars2.json")
	require.NoError(t, err)

	loaded, err := entries.Load("/tmp/vars2.json")
	require.NoError(t, err)
	require.True(t, loaded.Encrypted())
	require.Equal(t, "keychain", loaded.Encryption.Mode)
}

func TestEncryptAlreadyEncrypted(t *testing.T) {
	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(nil, nil))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	file := &entries.File{
		Encryption: entries.EncryptionHeader{Mode: "password", Salt: "x", Verifier: "y"},
		Entries:    []entries.Entry{{Name: "FOO", Value: "enc:v1:xxx"}},
	}
	seedVariablesFile(t, "/tmp/encrypted.json", file)

	err := Run(ctx, "/tmp/encrypted.json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "already encrypted")
}

func TestEncryptNoEntries(t *testing.T) {
	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(nil, nil))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	file := &entries.File{Entries: []entries.Entry{}}
	seedVariablesFile(t, "/tmp/empty.json", file)

	err := Run(ctx, "/tmp/empty.json")
	require.ErrorIs(t, err, ErrNoEntries)
}
