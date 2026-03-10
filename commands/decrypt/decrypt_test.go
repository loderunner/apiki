package decrypt

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

func seedEncryptedFile(t *testing.T, path string, password string) *entries.File {
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
	return file
}

func TestDecryptPasswordMode(t *testing.T) {
	seedEncryptedFile(t, "/tmp/enc.json", "secret")

	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter([]string{"secret"}, []string{"yes"}))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	err := Run(ctx, "/tmp/enc.json")
	require.NoError(t, err)

	loaded, err := entries.Load("/tmp/enc.json")
	require.NoError(t, err)
	require.False(t, loaded.Encrypted())
	require.Equal(t, "bar", loaded.Entries[0].Value)
	require.Equal(t, "qux", loaded.Entries[1].Value)
}

func TestDecryptKeychainMode(t *testing.T) {
	mockKC := &testutil.MockKeychain{}
	file := &entries.File{
		Entries: []entries.Entry{{Name: "FOO", Value: "bar"}},
	}
	key, err := file.SetPasswordMode("secret")
	require.NoError(t, err)
	err = file.EncryptValues(key)
	require.NoError(t, err)
	file.SetKeychainMode()
	err = mockKC.Store(key)
	require.NoError(t, err)
	err = entries.Save("/tmp/kc.json", file)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(nil, []string{"yes"}))
	ctx = keychain.WithKeychain(ctx, mockKC)

	err = Run(ctx, "/tmp/kc.json")
	require.NoError(t, err)

	loaded, err := entries.Load("/tmp/kc.json")
	require.NoError(t, err)
	require.False(t, loaded.Encrypted())
}

func TestDecryptNotEncrypted(t *testing.T) {
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

func TestDecryptNoEntries(t *testing.T) {
	file := &entries.File{
		Encryption: entries.EncryptionHeader{Mode: "password", Salt: "x", Verifier: "y"},
		Entries:    []entries.Entry{},
	}
	err := entries.Save("/tmp/empty.json", file)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(nil, nil))
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	err = Run(ctx, "/tmp/empty.json")
	require.ErrorIs(t, err, ErrNoEntries)
}
