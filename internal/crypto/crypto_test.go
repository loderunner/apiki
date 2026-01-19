package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSalt(t *testing.T) {
	t.Run("generates correct size", func(t *testing.T) {
		salt, err := GenerateSalt()
		require.NoError(t, err)
		require.Len(t, salt, SaltSize)
	})

	t.Run("generates unique salts", func(t *testing.T) {
		salt1, err1 := GenerateSalt()
		require.NoError(t, err1)

		salt2, err2 := GenerateSalt()
		require.NoError(t, err2)

		require.NotEqual(t, salt1, salt2, "salts should be unique")
	})
}

func TestDeriveKey(t *testing.T) {
	t.Run(
		"derives different keys from different passwords",
		func(t *testing.T) {
			salt := []byte("test-salt-16-b")

			key1 := DeriveKey("password1", salt)
			key2 := DeriveKey("password2", salt)

			require.NotEqual(
				t,
				key1,
				key2,
				"different passwords should produce different keys",
			)
		},
	)

	t.Run("derives different keys from different salts", func(t *testing.T) {
		password := "test-password"

		key1 := DeriveKey(password, []byte("salt1-16-bytes"))
		key2 := DeriveKey(password, []byte("salt2-16-bytes"))

		require.NotEqual(
			t,
			key1,
			key2,
			"different salts should produce different keys",
		)
	})

	t.Run("produces 32-byte key", func(t *testing.T) {
		key := DeriveKey("password", []byte("test-salt-16-b"))
		require.Len(t, key, KeySize)
	})
}

func TestComputeVerifier(t *testing.T) {
	t.Run(
		"produces different verifiers for different keys",
		func(t *testing.T) {
			salt := []byte("test-salt-16-b")

			verifier1 := ComputeVerifier(
				[]byte("key1-32-bytes-long-exactly!!"),
				salt,
			)
			verifier2 := ComputeVerifier(
				[]byte("key2-32-bytes-long-exactly!!"),
				salt,
			)

			require.NotEqual(t, verifier1, verifier2)
		},
	)

	t.Run(
		"produces different verifiers for different salts",
		func(t *testing.T) {
			key := []byte("test-key-32-bytes-long-exactly!")

			verifier1 := ComputeVerifier(key, []byte("salt1-16-bytes"))
			verifier2 := ComputeVerifier(key, []byte("salt2-16-bytes"))

			require.NotEqual(t, verifier1, verifier2)
		},
	)
}

func TestVerifyPassword(t *testing.T) {
	correctPassword := "correct-password"
	salt, err := GenerateSalt()
	require.NoError(t, err)

	key := DeriveKey(correctPassword, salt)
	verifier := ComputeVerifier(key, salt)

	t.Run("verifies correct password", func(t *testing.T) {
		verified := VerifyPassword("correct-password", salt, verifier)
		require.True(t, verified, "correct password should verify")
	})

	t.Run("rejects incorrect password", func(t *testing.T) {
		verified := VerifyPassword("wrong-password", salt, verifier)
		require.False(t, verified, "incorrect password should not verify")
	})

	t.Run("rejects password with wrong salt", func(t *testing.T) {
		verified := VerifyPassword(
			correctPassword,
			[]byte("wrong-salt-16-b"),
			verifier,
		)
		require.False(t, verified, "password with wrong salt should not verify")
	})
}

func TestGenerateKey(t *testing.T) {
	t.Run("generates correct size", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)
		require.Len(t, key, KeySize)
	})

	t.Run("generates unique keys", func(t *testing.T) {
		key1, err1 := GenerateKey()
		require.NoError(t, err1)

		key2, err2 := GenerateKey()
		require.NoError(t, err2)

		require.NotEqual(t, key1, key2, "keys should be unique")
	})
}

func TestIsEncrypted(t *testing.T) {
	t.Run("returns true for encrypted values", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)

		encrypted, err := Encrypt(key, "secret")
		require.NoError(t, err)

		require.True(t, IsEncrypted(encrypted))
	})

	t.Run("returns false for plaintext values", func(t *testing.T) {
		testCases := []string{
			"",
			"plaintext",
			"enc:",
			"enc:v",
			"enc:v1",
			"enc:v2:something",
			"some-random-value",
		}

		for _, value := range testCases {
			require.False(t, IsEncrypted(value), "expected false for %q", value)
		}
	})

	t.Run("returns true for prefix only", func(t *testing.T) {
		// Edge case: just the prefix with no payload
		require.True(t, IsEncrypted("enc:v1:"))
	})
}

func TestEncrypt(t *testing.T) {
	t.Run("encrypts plaintext successfully", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)

		plaintext := "secret-value-123"
		encrypted, err := Encrypt(key, plaintext)
		require.NoError(t, err)
		require.NotEmpty(t, encrypted)
		require.Contains(
			t,
			encrypted,
			prefix,
			"encrypted value should have prefix",
		)
	})

	t.Run("rejects invalid key size", func(t *testing.T) {
		invalidKey := []byte("too-short")
		_, err := Encrypt(invalidKey, "plaintext")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid key size")
	})
}

func TestDecrypt(t *testing.T) {
	t.Run("round-trip encryption", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)

		testCases := []string{
			"",
			"a",
			"short",
			"medium-length-string",
			"very-long-string-" + string(make([]byte, 1000)),
			"special-chars-!@#$%^&*()",
			"unicode-测试-🚀",
			"newlines\nand\ttabs",
		}

		for _, plaintext := range testCases {
			encrypted, err := Encrypt(key, plaintext)
			require.NoError(
				t,
				err,
				"encryption should succeed for %q",
				plaintext,
			)

			decrypted, err := Decrypt(key, encrypted)
			require.NoError(
				t,
				err,
				"decryption should succeed for %q",
				plaintext,
			)
			require.Equal(
				t,
				plaintext,
				decrypted,
				"round-trip should preserve %q",
				plaintext,
			)
		}
	})

	t.Run("rejects wrong key", func(t *testing.T) {
		key1, err1 := GenerateKey()
		require.NoError(t, err1)

		key2, err2 := GenerateKey()
		require.NoError(t, err2)

		plaintext := "secret"
		encrypted, err := Encrypt(key1, plaintext)
		require.NoError(t, err)

		_, err = Decrypt(key2, encrypted)
		require.Error(t, err, "decryption with wrong key should fail")
		require.Contains(t, err.Error(), "failed to decrypt")
	})

	t.Run("rejects invalid prefix", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)

		_, err = Decrypt(key, "invalid-prefix:data")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid encryption format")
	})

	t.Run("rejects invalid base64", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)

		invalidBase64 := prefix + "not-valid-base64!!!"
		_, err = Decrypt(key, invalidBase64)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decode base64")
	})

	t.Run("rejects invalid key size", func(t *testing.T) {
		invalidKey := []byte("too-short")
		encrypted := prefix + "dGVzdC1kYXRhLWhlcmU="
		_, err := Decrypt(invalidKey, encrypted)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid key size")
	})
}

func TestEncryptDecryptIntegration(t *testing.T) {
	t.Run("multiple encrypt decrypt cycles", func(t *testing.T) {
		key, err := GenerateKey()
		require.NoError(t, err)

		prevEncrypted := ""
		plaintext := "test-value"
		for range 10 {
			encrypted, err := Encrypt(key, plaintext)
			require.NoError(t, err)
			require.NotEqual(
				t,
				prevEncrypted,
				encrypted,
				"encrypted value should be different on each cycle",
			)

			decrypted, err := Decrypt(key, encrypted)
			require.NoError(t, err)
			require.Equal(t, plaintext, decrypted)

			prevEncrypted = encrypted
		}
	})
}
