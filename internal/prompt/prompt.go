package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ReadPassword reads a password from stdin with hidden input.
// Prints the prompt to stderr and returns the password.
func ReadPassword(prompt string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s", prompt)

	// Read password from stdin
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	return string(password), nil
}

// ReadChoice reads a single character choice from stdin.
// Prints the prompt to stderr and returns the selected value.
// The choices map maps single characters (case-insensitive) to their values.
// For example, map['p']="password", map['k']="keychain" for p/k choice.
func ReadChoice(prompt string, choices map[rune]string) (string, error) {
	return ReadChoiceWithDefault(prompt, choices, "")
}

// ReadChoiceWithDefault reads a single character choice from stdin.
// If the user presses Enter without typing a character, the default value
// is returned. Pass an empty string for no default.
func ReadChoiceWithDefault(
	prompt string,
	choices map[rune]string,
	defaultValue string,
) (string, error) {
	fmt.Fprintf(os.Stderr, "%s", prompt)

	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()
	if err != nil {
		return "", fmt.Errorf("failed to read choice: %w", err)
	}

	// Handle Enter key as default
	if char == '\n' || char == '\r' {
		if defaultValue != "" {
			return defaultValue, nil
		}
		return "", fmt.Errorf("no choice entered")
	}

	// Normalize to lowercase for case-insensitive matching
	charLower := rune(strings.ToLower(string(char))[0])

	if value, ok := choices[charLower]; ok {
		return value, nil
	}

	return "", fmt.Errorf("invalid choice: %c", char)
}
