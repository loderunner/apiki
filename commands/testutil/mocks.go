package testutil

import (
	"errors"
	"sync"
)

var (
	keychainRetrieveErr = errors.New("mock keychain retrieve failed")
	promptReadErr       = errors.New("mock prompter: no more values")
)

// MockKeychain implements keychain.Keychain for testing.
type MockKeychain struct {
	mu   sync.Mutex
	key  []byte
	fail bool // if true, Retrieve returns error
}

// Store stores the key in memory.
func (m *MockKeychain) Store(key []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.key = make([]byte, len(key))
	copy(m.key, key)
	return nil
}

// Retrieve returns the stored key.
func (m *MockKeychain) Retrieve() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail || m.key == nil {
		return nil, keychainRetrieveErr
	}
	out := make([]byte, len(m.key))
	copy(out, m.key)
	return out, nil
}

// Delete clears the stored key.
func (m *MockKeychain) Delete() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.key = nil
	return nil
}

// SetFail makes Retrieve return an error.
func (m *MockKeychain) SetFail(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fail = fail
}

// MockPrompter implements prompt.Prompter for testing.
type MockPrompter struct {
	passwords []string
	choices   []string
	passwordI int
	choiceI   int
}

// NewMockPrompter creates a MockPrompter that returns the given passwords
// and choices in order. For ReadChoice/ReadChoiceWithDefault, each call
// consumes the next choice.
func NewMockPrompter(passwords, choices []string) *MockPrompter {
	return &MockPrompter{
		passwords: passwords,
		choices:   choices,
	}
}

// ReadPassword returns the next password from the list.
func (m *MockPrompter) ReadPassword(prompt string) (string, error) {
	if m.passwordI >= len(m.passwords) {
		return "", promptReadErr
	}
	s := m.passwords[m.passwordI]
	m.passwordI++
	return s, nil
}

// ReadChoice returns the next choice from the list.
func (m *MockPrompter) ReadChoice(
	prompt string,
	choices map[rune]string,
) (string, error) {
	if m.choiceI >= len(m.choices) {
		return "", promptReadErr
	}
	s := m.choices[m.choiceI]
	m.choiceI++
	return s, nil
}

// ReadChoiceWithDefault returns the next choice from the list.
func (m *MockPrompter) ReadChoiceWithDefault(
	prompt string,
	choices map[rune]string,
	defaultValue string,
) (string, error) {
	return m.ReadChoice(prompt, choices)
}
