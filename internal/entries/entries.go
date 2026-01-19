package entries

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/loderunner/apiki/internal/crypto"
	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

// File represents the on-disk structure (JSON file)
type File struct {
	Encryption EncryptionHeader `json:"encryption"`
	Entries    []Entry          `json:"entries"`
}

// EncryptionHeader holds encryption metadata.
// Zero value means unencrypted (Mode == "").
type EncryptionHeader struct {
	// "password", or "keychain"
	Mode string `json:"mode,omitempty"`
	// base64, only for password mode
	Salt string `json:"salt,omitempty"`
	// base64, only for password mode
	Verifier string `json:"verifier,omitempty"`
}

// Enabled returns true if encryption is configured.
func (h EncryptionHeader) Enabled() bool {
	return h.Mode != ""
}

// Entry represents an environment variable entry.
// This is the serializable version (no Selected, SourceFile fields).
type Entry struct {
	Name  string `json:"name"`
	Value string `json:"value"` // plaintext or "enc:v1:..." ciphertext
	Label string `json:"label,omitempty"`
}

// Load reads the file from disk and parses it into memory.
func Load(path string) (*File, error) {
	dir := filepath.Dir(path)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := afero.ReadFile(fs, path)
	if err != nil {
		if errors.Is(err, afero.ErrFileNotFound) {
			return &File{
				Encryption: EncryptionHeader{},
				Entries:    []Entry{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return &File{
			Encryption: EncryptionHeader{},
			Entries:    []Entry{},
		}, nil
	}

	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &file, nil
}

// Save serializes the in-memory model and writes it to disk.
func Save(path string, f *File) error {
	dir := filepath.Dir(path)
	if err := fs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := afero.WriteFile(fs, path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Encrypted returns true if encryption is enabled.
func (f *File) Encrypted() bool {
	return f.Encryption.Enabled()
}

// EncryptValues encrypts all entry values in place using the given key.
// Returns an error if any value is already encrypted.
func (f *File) EncryptValues(key []byte) error {
	for i := range f.Entries {
		if crypto.IsEncrypted(f.Entries[i].Value) {
			return fmt.Errorf(
				"variable %q is already encrypted",
				f.Entries[i].Name,
			)
		}

		encrypted, err := crypto.Encrypt(key, f.Entries[i].Value)
		if err != nil {
			return fmt.Errorf(
				"failed to encrypt variable %q: %w",
				f.Entries[i].Name,
				err,
			)
		}
		f.Entries[i].Value = encrypted
	}
	return nil
}

// DecryptValues decrypts all entry values in place using the given key.
// Returns an error if any value is not encrypted.
func (f *File) DecryptValues(key []byte) error {
	for i := range f.Entries {
		if !crypto.IsEncrypted(f.Entries[i].Value) {
			return fmt.Errorf("variable %q is not encrypted", f.Entries[i].Name)
		}

		decrypted, err := crypto.Decrypt(key, f.Entries[i].Value)
		if err != nil {
			return fmt.Errorf(
				"failed to decrypt variable %q: %w",
				f.Entries[i].Name,
				err,
			)
		}
		f.Entries[i].Value = decrypted
	}
	return nil
}

// Clone returns a deep copy of the file.
func (f *File) Clone() *File {
	clone := &File{
		Encryption: f.Encryption,
		Entries:    make([]Entry, len(f.Entries)),
	}
	copy(clone.Entries, f.Entries)
	return clone
}

// VerifyPassword verifies a password against the encryption header.
// Returns the derived key if verification succeeds.
func (f *File) VerifyPassword(password string) ([]byte, error) {
	if f.Encryption.Mode != "password" {
		return nil, errors.New("file is not password-protected")
	}

	salt, err := base64.StdEncoding.DecodeString(f.Encryption.Salt)
	if err != nil {
		return nil, fmt.Errorf("invalid salt: %w", err)
	}

	verifier, err := base64.StdEncoding.DecodeString(f.Encryption.Verifier)
	if err != nil {
		return nil, fmt.Errorf("invalid verifier: %w", err)
	}

	key := crypto.DeriveKey(password, salt)
	if !crypto.VerifyPassword(password, salt, verifier) {
		return nil, errors.New("wrong password")
	}

	return key, nil
}

// SetPasswordMode configures password-based encryption.
// Returns the derived encryption key.
func (f *File) SetPasswordMode(password string) ([]byte, error) {
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := crypto.DeriveKey(password, salt)
	verifier := crypto.ComputeVerifier(key, salt)

	f.Encryption = EncryptionHeader{
		Mode:     "password",
		Salt:     base64.StdEncoding.EncodeToString(salt),
		Verifier: base64.StdEncoding.EncodeToString(verifier),
	}

	return key, nil
}

// SetKeychainMode configures keychain-based encryption.
func (f *File) SetKeychainMode() {
	f.Encryption = EncryptionHeader{
		Mode: "keychain",
	}
}

// ClearEncryption removes encryption configuration.
func (f *File) ClearEncryption() {
	f.Encryption = EncryptionHeader{}
}
