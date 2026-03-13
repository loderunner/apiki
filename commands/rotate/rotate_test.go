package rotate

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

func seedEncryptedFile(t *testing.T, path string, password string) {
	t.Helper()
	file := &entries.File{
		Entries: []entries.Entry{
			{Name: "FOO", Value: "bar"},
			{Name: "BAZ", Value: "qux"},
		},
	}
	key, err := file.SetPasswordMode(password)
	require.NoError(t, err)
	err = file.EncryptValues(key)
	require.NoError(t, err)
	err = entries.Save(path, file)
	require.NoError(t, err)
}

func TestRotatePasswordToPassword(t *testing.T) {
	seedEncryptedFile(t, "/tmp/rot.json", "oldpass")

	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(
		[]string{"oldpass", "newpass", "newpass"},
		[]string{"password", "password"},
	))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	err := Run(ctx, "/tmp/rot.json")
	require.NoError(t, err)

	loaded, err := entries.Load("/tmp/rot.json")
	require.NoError(t, err)
	require.True(t, loaded.Encrypted())
	require.Equal(t, "password", loaded.Encryption.Mode)
	// Verify we can decrypt with new password
	key, err := loaded.VerifyPassword("newpass")
	require.NoError(t, err)
	err = loaded.DecryptValues(key)
	require.NoError(t, err)
	require.Equal(t, "bar", loaded.Entries[0].Value)
}

func TestRotateNotEncrypted(t *testing.T) {
	file := &entries.File{
		Entries: []entries.Entry{{Name: "FOO", Value: "bar"}},
	}
	err := entries.Save("/tmp/plain.json", file)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(nil, nil))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	err = Run(ctx, "/tmp/plain.json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not encrypted")
}

func TestRotateNoEntries(t *testing.T) {
	file := &entries.File{
		Encryption: entries.EncryptionHeader{
			Mode:     "password",
			Salt:     "x",
			Verifier: "y",
		},
		Entries: []entries.Entry{},
	}
	err := entries.Save("/tmp/empty.json", file)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(nil, nil))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	err = Run(ctx, "/tmp/empty.json")
	require.ErrorIs(t, err, ErrNoEntries)
}
