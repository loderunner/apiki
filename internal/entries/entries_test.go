package entries

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/loderunner/apiki/internal/crypto"
)

func TestMain(m *testing.M) {
	// Use an in-memory filesystem for testing
	fs = afero.NewMemMapFs()
	os.Exit(m.Run())
}

func TestLoad(t *testing.T) {
	t.Run("returns empty file when file does not exist", func(t *testing.T) {
		file, err := Load("/nonexistent/file.json")
		require.NoError(t, err)
		require.NotNil(t, file)
		require.False(t, file.Encrypted())
		require.Empty(t, file.Entries)
	})

	t.Run("returns empty file when file is empty", func(t *testing.T) {
		path := "/test/empty.json"
		err := afero.WriteFile(fs, path, []byte{}, 0o644)
		require.NoError(t, err)

		file, err := Load(path)
		require.NoError(t, err)
		require.NotNil(t, file)
		require.False(t, file.Encrypted())
		require.Empty(t, file.Entries)
	})

	t.Run("loads valid file with entries", func(t *testing.T) {
		path := "/test/valid.json"
		expectedFile := &File{
			Encryption: EncryptionHeader{},
			Entries: []Entry{
				{Name: "VAR1", Value: "value1"},
				{Name: "VAR2", Value: "value2", Label: "test label"},
			},
		}

		data, err := json.MarshalIndent(expectedFile, "", "  ")
		require.NoError(t, err)
		err = afero.WriteFile(fs, path, data, 0o644)
		require.NoError(t, err)

		file, err := Load(path)
		require.NoError(t, err)
		require.NotNil(t, file)
		require.Equal(t, expectedFile.Entries, file.Entries)
		require.False(t, file.Encrypted())
	})

	t.Run("loads file with encryption header", func(t *testing.T) {
		path := "/test/encrypted.json"
		expectedFile := &File{
			Encryption: EncryptionHeader{
				Mode:     "password",
				Salt:     "dGVzdC1zYWx0",
				Verifier: "dGVzdC12ZXJpZmllcg==",
			},
			Entries: []Entry{
				{Name: "VAR1", Value: "enc:v1:encrypted"},
			},
		}

		data, err := json.MarshalIndent(expectedFile, "", "  ")
		require.NoError(t, err)
		err = afero.WriteFile(fs, path, data, 0o644)
		require.NoError(t, err)

		file, err := Load(path)
		require.NoError(t, err)
		require.NotNil(t, file)
		require.True(t, file.Encrypted())
		require.Equal(t, "password", file.Encryption.Mode)
		require.Equal(t, expectedFile.Entries, file.Entries)
	})

	t.Run("creates directory if it does not exist", func(t *testing.T) {
		path := "/new/dir/file.json"
		file, err := Load(path)
		require.NoError(t, err)
		require.NotNil(t, file)

		exists, err := afero.DirExists(fs, "/new/dir")
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		path := "/test/invalid.json"
		err := afero.WriteFile(fs, path, []byte("{invalid json}"), 0o644)
		require.NoError(t, err)

		_, err = Load(path)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse JSON")
	})
}

func TestSave(t *testing.T) {
	t.Run("saves file successfully", func(t *testing.T) {
		path := "/test/save.json"
		file := &File{
			Encryption: EncryptionHeader{},
			Entries: []Entry{
				{Name: "VAR1", Value: "value1"},
				{Name: "VAR2", Value: "value2"},
			},
		}

		err := Save(path, file)
		require.NoError(t, err)

		data, err := afero.ReadFile(fs, path)
		require.NoError(t, err)

		var loaded File
		err = json.Unmarshal(data, &loaded)
		require.NoError(t, err)
		require.Equal(t, file.Entries, loaded.Entries)
	})

	t.Run("creates directory if it does not exist", func(t *testing.T) {
		path := "/new/dir/save.json"
		file := &File{
			Encryption: EncryptionHeader{},
			Entries:    []Entry{},
		}

		err := Save(path, file)
		require.NoError(t, err)

		exists, err := afero.DirExists(fs, "/new/dir")
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("saves file with encryption header", func(t *testing.T) {
		path := "/test/encrypted.json"
		file := &File{
			Encryption: EncryptionHeader{
				Mode:     "password",
				Salt:     "dGVzdC1zYWx0",
				Verifier: "dGVzdC12ZXJpZmllcg==",
			},
			Entries: []Entry{
				{Name: "VAR1", Value: "enc:v1:encrypted"},
			},
		}

		err := Save(path, file)
		require.NoError(t, err)

		loaded, err := Load(path)
		require.NoError(t, err)
		require.Equal(t, file.Encryption, loaded.Encryption)
		require.Equal(t, file.Entries, loaded.Entries)
	})
}

func TestEncrypted(t *testing.T) {
	t.Run("returns false for unencrypted file", func(t *testing.T) {
		file := &File{
			Encryption: EncryptionHeader{},
		}
		require.False(t, file.Encrypted())
	})

	t.Run("returns true for password-encrypted file", func(t *testing.T) {
		file := &File{
			Encryption: EncryptionHeader{Mode: "password"},
		}
		require.True(t, file.Encrypted())
	})

	t.Run("returns true for keychain-encrypted file", func(t *testing.T) {
		file := &File{
			Encryption: EncryptionHeader{Mode: "keychain"},
		}
		require.True(t, file.Encrypted())
	})
}

func TestEncryptValues(t *testing.T) {
	t.Run("encrypts plaintext values", func(t *testing.T) {
		key, err := crypto.GenerateKey()
		require.NoError(t, err)

		file := &File{
			Entries: []Entry{
				{Name: "VAR1", Value: "plaintext1"},
				{Name: "VAR2", Value: "plaintext2"},
			},
		}

		err = file.EncryptValues(key)
		require.NoError(t, err)

		require.True(
			t,
			crypto.IsEncrypted(file.Entries[0].Value),
			"VAR1 should be encrypted",
		)
		require.True(
			t,
			crypto.IsEncrypted(file.Entries[1].Value),
			"VAR2 should be encrypted",
		)
	})

	t.Run("errors on already encrypted values", func(t *testing.T) {
		key, err := crypto.GenerateKey()
		require.NoError(t, err)

		encrypted, err := crypto.Encrypt(key, "already-encrypted-value")
		require.NoError(t, err)

		file := &File{
			Entries: []Entry{
				{Name: "VAR1", Value: encrypted},
				{Name: "VAR2", Value: "plaintext"},
			},
		}

		err = file.EncryptValues(key)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already encrypted")
	})

	t.Run("returns error for invalid key", func(t *testing.T) {
		invalidKey := []byte("too-short")
		file := &File{
			Entries: []Entry{
				{Name: "VAR1", Value: "plaintext"},
			},
		}

		err := file.EncryptValues(invalidKey)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to encrypt variable")
	})
}

func TestDecryptValues(t *testing.T) {
	t.Run("decrypts encrypted values", func(t *testing.T) {
		key, err := crypto.GenerateKey()
		require.NoError(t, err)

		plaintext1 := "secret1"
		plaintext2 := "secret2"

		encrypted1, err := crypto.Encrypt(key, plaintext1)
		require.NoError(t, err)
		encrypted2, err := crypto.Encrypt(key, plaintext2)
		require.NoError(t, err)

		file := &File{
			Entries: []Entry{
				{Name: "VAR1", Value: encrypted1},
				{Name: "VAR2", Value: encrypted2},
			},
		}

		// Store original encrypted values for comparison
		original1 := file.Entries[0].Value
		original2 := file.Entries[1].Value

		err = file.DecryptValues(key)
		require.NoError(t, err)

		// Values should be decrypted (different from encrypted)
		require.NotEqual(t, original1, file.Entries[0].Value)
		require.NotEqual(t, original2, file.Entries[1].Value)
		require.Equal(t, plaintext1, file.Entries[0].Value)
		require.Equal(t, plaintext2, file.Entries[1].Value)
	})

	t.Run("errors on plaintext values", func(t *testing.T) {
		key, err := crypto.GenerateKey()
		require.NoError(t, err)

		file := &File{
			Entries: []Entry{
				{Name: "VAR1", Value: "plaintext-value"},
			},
		}

		err = file.DecryptValues(key)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not encrypted")
	})

	t.Run("returns error for wrong key", func(t *testing.T) {
		key1, err := crypto.GenerateKey()
		require.NoError(t, err)
		key2, err := crypto.GenerateKey()
		require.NoError(t, err)

		encrypted, err := crypto.Encrypt(key1, "secret")
		require.NoError(t, err)

		file := &File{
			Entries: []Entry{
				{Name: "VAR1", Value: encrypted},
			},
		}

		err = file.DecryptValues(key2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decrypt variable")
	})
}

func TestClone(t *testing.T) {
	t.Run("creates deep copy", func(t *testing.T) {
		original := &File{
			Encryption: EncryptionHeader{
				Mode:     "password",
				Salt:     "dGVzdC1zYWx0",
				Verifier: "dGVzdC12ZXJpZmllcg==",
			},
			Entries: []Entry{
				{Name: "VAR1", Value: "value1"},
				{Name: "VAR2", Value: "value2", Label: "label"},
			},
		}

		clone := original.Clone()

		require.Equal(t, original.Encryption, clone.Encryption)
		require.Equal(t, original.Entries, clone.Entries)

		clone.Entries[0].Value = "modified"
		require.NotEqual(t, original.Entries[0].Value, clone.Entries[0].Value)
	})

	t.Run("handles empty file", func(t *testing.T) {
		original := &File{
			Encryption: EncryptionHeader{},
			Entries:    []Entry{},
		}

		clone := original.Clone()
		require.Equal(t, original, clone)
		require.Empty(t, clone.Entries)
	})
}

func TestVerifyPassword(t *testing.T) {
	t.Run("verifies correct password", func(t *testing.T) {
		password := "test-password"
		file := &File{}
		key, err := file.SetPasswordMode(password)
		require.NoError(t, err)

		verifiedKey, err := file.VerifyPassword(password)
		require.NoError(t, err)
		require.Equal(t, key, verifiedKey)
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		password := "test-password"
		file := &File{}
		_, err := file.SetPasswordMode(password)
		require.NoError(t, err)

		_, err = file.VerifyPassword("wrong-password")
		require.Error(t, err)
		require.Contains(t, err.Error(), "wrong password")
	})

	t.Run("returns error for non-password mode", func(t *testing.T) {
		file := &File{
			Encryption: EncryptionHeader{Mode: "keychain"},
		}

		_, err := file.VerifyPassword("password")
		require.Error(t, err)
		require.Contains(t, err.Error(), "file is not password-protected")
	})

	t.Run("returns error for invalid salt", func(t *testing.T) {
		file := &File{
			Encryption: EncryptionHeader{
				Mode:     "password",
				Salt:     "invalid-base64!!!",
				Verifier: "dGVzdC12ZXJpZmllcg==",
			},
		}

		_, err := file.VerifyPassword("password")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid salt")
	})

	t.Run("returns error for invalid verifier", func(t *testing.T) {
		salt, err := crypto.GenerateSalt()
		require.NoError(t, err)

		file := &File{
			Encryption: EncryptionHeader{
				Mode:     "password",
				Salt:     base64.StdEncoding.EncodeToString(salt),
				Verifier: "invalid-base64!!!",
			},
		}

		_, err = file.VerifyPassword("password")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid verifier")
	})
}

func TestSetPasswordMode(t *testing.T) {
	t.Run("sets password mode and returns key", func(t *testing.T) {
		file := &File{}
		password := "test-password"

		key, err := file.SetPasswordMode(password)
		require.NoError(t, err)
		require.NotNil(t, key)
		require.Len(t, key, crypto.KeySize)

		require.Equal(t, "password", file.Encryption.Mode)
		require.NotEmpty(t, file.Encryption.Salt)
		require.NotEmpty(t, file.Encryption.Verifier)
	})

	t.Run("produces different salts on each call", func(t *testing.T) {
		file1 := &File{}
		file2 := &File{}
		password := "test-password"

		_, err1 := file1.SetPasswordMode(password)
		require.NoError(t, err1)
		_, err2 := file2.SetPasswordMode(password)
		require.NoError(t, err2)

		require.NotEqual(t, file1.Encryption.Salt, file2.Encryption.Salt)
	})

	t.Run("allows password verification after setting", func(t *testing.T) {
		file := &File{}
		password := "test-password"

		key, err := file.SetPasswordMode(password)
		require.NoError(t, err)

		verifiedKey, err := file.VerifyPassword(password)
		require.NoError(t, err)
		require.Equal(t, key, verifiedKey)
	})
}

func TestSetKeychainMode(t *testing.T) {
	t.Run("sets keychain mode", func(t *testing.T) {
		file := &File{
			Encryption: EncryptionHeader{
				Mode:     "password",
				Salt:     "dGVzdC1zYWx0",
				Verifier: "dGVzdC12ZXJpZmllcg==",
			},
		}

		file.SetKeychainMode()

		require.Equal(t, "keychain", file.Encryption.Mode)
		require.Empty(t, file.Encryption.Salt)
		require.Empty(t, file.Encryption.Verifier)
	})

	t.Run("overwrites existing encryption", func(t *testing.T) {
		file := &File{}
		_, err := file.SetPasswordMode("password")
		require.NoError(t, err)

		file.SetKeychainMode()

		require.Equal(t, "keychain", file.Encryption.Mode)
		require.Empty(t, file.Encryption.Salt)
		require.Empty(t, file.Encryption.Verifier)
	})
}

func TestClearEncryption(t *testing.T) {
	t.Run("clears password encryption", func(t *testing.T) {
		file := &File{}
		_, err := file.SetPasswordMode("password")
		require.NoError(t, err)

		file.ClearEncryption()

		require.False(t, file.Encrypted())
		require.Empty(t, file.Encryption.Mode)
		require.Empty(t, file.Encryption.Salt)
		require.Empty(t, file.Encryption.Verifier)
	})

	t.Run("clears keychain encryption", func(t *testing.T) {
		file := &File{}
		file.SetKeychainMode()

		file.ClearEncryption()

		require.False(t, file.Encrypted())
		require.Empty(t, file.Encryption.Mode)
	})
}

func TestEncryptionHeaderEnabled(t *testing.T) {
	t.Run("returns false for zero value", func(t *testing.T) {
		header := EncryptionHeader{}
		require.False(t, header.Enabled())
	})

	t.Run("returns true for password mode", func(t *testing.T) {
		header := EncryptionHeader{Mode: "password"}
		require.True(t, header.Enabled())
	})

	t.Run("returns true for keychain mode", func(t *testing.T) {
		header := EncryptionHeader{Mode: "keychain"}
		require.True(t, header.Enabled())
	})
}
