package prompt

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Prompter defines the interface for reading user input.
type Prompter interface {
	ReadPassword(prompt string) (string, error)
	ReadChoice(prompt string, choices map[rune]string) (string, error)
	ReadChoiceWithDefault(
		prompt string,
		choices map[rune]string,
		defaultValue string,
	) (string, error)
}

type contextKey struct{}

// WithPrompter returns a context with the given prompter for injection.
func WithPrompter(ctx context.Context, p Prompter) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

func fromContext(ctx context.Context) Prompter {
	if p, ok := ctx.Value(contextKey{}).(Prompter); ok {
		return p
	}
	return terminalPrompter{}
}

type terminalPrompter struct{}

func (terminalPrompter) ReadPassword(promptStr string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s", promptStr)

	// Read password from stdin
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	return string(password), nil
}

func (t terminalPrompter) ReadChoice(
	promptStr string,
	choices map[rune]string,
) (string, error) {
	return t.ReadChoiceWithDefault(promptStr, choices, "")
}

func (terminalPrompter) ReadChoiceWithDefault(
	promptStr string,
	choices map[rune]string,
	defaultValue string,
) (string, error) {
	fmt.Fprintf(os.Stderr, "%s", promptStr)

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

// ReadPassword reads a password using the prompter from context.
// Prints the prompt to stderr and returns the password.
func ReadPassword(ctx context.Context, promptStr string) (string, error) {
	return fromContext(ctx).ReadPassword(promptStr)
}

// ReadChoice reads a single character choice using the prompter from context.
// The choices map maps single characters (case-insensitive) to their values.
// For example, map['p']="password", map['k']="keychain" for p/k choice.
func ReadChoice(
	ctx context.Context,
	promptStr string,
	choices map[rune]string,
) (string, error) {
	return fromContext(ctx).ReadChoice(promptStr, choices)
}

// ReadChoiceWithDefault reads a single character choice using the prompter from
// context.
// If the user presses Enter without typing a character, the default value
// is returned. Pass an empty string for no default.
func ReadChoiceWithDefault(
	ctx context.Context,
	promptStr string,
	choices map[rune]string,
	defaultValue string,
) (string, error) {
	return fromContext(
		ctx,
	).ReadChoiceWithDefault(promptStr, choices, defaultValue)
}
