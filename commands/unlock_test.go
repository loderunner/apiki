package commands

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

func TestUnlockPasswordMode(t *testing.T) {
	file := &entries.File{
		Entries: []entries.Entry{{Name: "FOO", Value: "bar"}},
	}
	key, err := file.SetPasswordMode("secret")
	require.NoError(t, err)
	err = file.EncryptValues(key)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = prompt.WithPrompter(
		ctx,
		testutil.NewMockPrompter([]string{"secret"}, nil),
	)
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	gotKey, err := Unlock(ctx, file)
	require.NoError(t, err)
	require.Len(t, gotKey, 32)
	err = file.DecryptValues(gotKey)
	require.NoError(t, err)
	require.Equal(t, "bar", file.Entries[0].Value)
}

func TestUnlockPasswordWrongPassword(t *testing.T) {
	file := &entries.File{
		Entries: []entries.Entry{{Name: "FOO", Value: "bar"}},
	}
	key, err := file.SetPasswordMode("secret")
	require.NoError(t, err)
	err = file.EncryptValues(key)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = prompt.WithPrompter(
		ctx,
		testutil.NewMockPrompter([]string{"wrong", "wrong"}, nil),
	)
	ctx = keychain.WithKeychain(ctx, &testutil.MockKeychain{})

	_, err = Unlock(ctx, file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "too many wrong password")
}

func TestUnlockKeychainMode(t *testing.T) {
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

	ctx := context.Background()
	ctx = prompt.WithPrompter(ctx, testutil.NewMockPrompter(nil, nil))
	ctx = keychain.WithKeychain(ctx, mockKC)

	gotKey, err := Unlock(ctx, file)
	require.NoError(t, err)
	require.Len(t, gotKey, 32)
}

func TestUnlockNotEncrypted(t *testing.T) {
	file := &entries.File{
		Entries: []entries.Entry{{Name: "FOO", Value: "bar"}},
	}

	ctx := context.Background()
	_, err := Unlock(ctx, file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not encrypted")
}
