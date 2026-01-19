package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/loderunner/apiki/cmd/decrypt"
	"github.com/loderunner/apiki/cmd/encrypt"
	"github.com/loderunner/apiki/cmd/rotate"
	"github.com/loderunner/apiki/internal/entries"
	"github.com/loderunner/apiki/internal/keychain"
	"github.com/loderunner/apiki/internal/prompt"
)

var version = "dev"

// variablesFile holds the value of the --variables-file flag.
var variablesFile string

func main() {
	rootCmd := &cobra.Command{
		Use:   "apiki",
		Short: "Environment variable manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			variablesPath, err := resolveVariablesFile(cmd)
			if err != nil {
				return fmt.Errorf("could not resolve variables file: %w", err)
			}
			output, err := run(variablesPath)
			if err != nil {
				return err
			}
			if output != "" {
				fmt.Printf("%s\n", output)
			}
			return nil
		},
	}

	// Persistent flag available to root and all subcommands
	rootCmd.PersistentFlags().StringVarP(
		&variablesFile,
		"variables-file", "f",
		"",
		"path to variables file (env: APIKI_FILE)",
	)

	// Redirect all Cobra output to stderr to avoid breaking eval
	rootCmd.SetOut(os.Stderr)
	rootCmd.SetErr(os.Stderr)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stderr, "apiki %s\n", strings.TrimSpace(version))
		},
	}

	encryptCmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt variable values",
		RunE: func(cmd *cobra.Command, args []string) error {
			variablesPath, err := resolveVariablesFile(cmd)
			if err != nil {
				return fmt.Errorf("could not resolve variables file: %w", err)
			}
			err = encrypt.Run(variablesPath)
			if errors.Is(err, encrypt.ErrNoEntries) {
				cmd.PrintErrln(err.Error())
				return nil
			}
			return err
		},
	}

	decryptCmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt variable values",
		RunE: func(cmd *cobra.Command, args []string) error {
			variablesPath, err := resolveVariablesFile(cmd)
			if err != nil {
				return fmt.Errorf("could not resolve variables file: %w", err)
			}
			err = decrypt.Run(variablesPath)
			if errors.Is(err, decrypt.ErrNoEntries) {
				cmd.PrintErrln(err.Error())
				return nil
			}
			return err
		},
	}

	rotateCmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate encryption key",
		RunE: func(cmd *cobra.Command, args []string) error {
			variablesPath, err := resolveVariablesFile(cmd)
			if err != nil {
				return fmt.Errorf("could not resolve variables file: %w", err)
			}
			err = rotate.Run(variablesPath)
			if errors.Is(err, rotate.ErrNoEntries) {
				cmd.PrintErrln(err.Error())
				return nil
			}
			return err
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)
	rootCmd.AddCommand(rotateCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// resolveVariablesFile determines the variables file path using the following
// priority:
//  1. --variables-file flag (if explicitly set)
//  2. APIKI_FILE environment variable
//  3. Default path (~/.apiki/variables.json)
func resolveVariablesFile(cmd *cobra.Command) (string, error) {
	// 1. Check if flag was explicitly set
	if cmd.Flags().Changed("variables-file") {
		return variablesFile, nil
	}

	// 2. Check environment variable
	if envPath := os.Getenv("APIKI_FILE"); envPath != "" {
		return envPath, nil
	}

	// 3. Fall back to default
	return DefaultEntriesPath()
}

func run(entriesPath string) (string, error) {
	// Load file (may be encrypted)
	file, err := entries.Load(entriesPath)
	if err != nil {
		return "", fmt.Errorf("could not load variables file: %w", err)
	}

	// Unlock if encrypted
	var encryptionKey []byte
	if file.Encrypted() {
		encryptionKey, err = unlockFile(file)
		if err != nil {
			return "", fmt.Errorf("failed to unlock file: %w", err)
		}

		// Decrypt values in memory
		if err := file.DecryptValues(encryptionKey); err != nil {
			return "", fmt.Errorf("failed to decrypt variables: %w", err)
		}
	}

	// Convert entries.File entries to TUI Entry format
	apikiEntries := make([]Entry, len(file.Entries))
	for i, e := range file.Entries {
		apikiEntries[i] = Entry{
			Name:       e.Name,
			Value:      e.Value,
			Label:      e.Label,
			Selected:   false,
			SourceFile: "",
		}
	}

	dotEnvEntries, err := LoadDotEnvEntries()
	if err != nil {
		return "", fmt.Errorf("could not load .env variables: %w", err)
	}

	// Combine apiki entries with .env entries (no deduplication)
	allEntries := append(apikiEntries, dotEnvEntries...)

	// Capture the environment state for all entry names at startup
	envSnapshot := captureEnvironment(allEntries)

	// Sync selection state with captured environment
	syncWithEnvironment(allEntries, envSnapshot)

	// Open TTY for TUI input/output, keeping stdout clean for shell commands
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("could not open /dev/tty: %w", err)
	}
	defer func() { _ = tty.Close() }()

	model := NewModel(file, entriesPath, encryptionKey, allEntries)
	p := tea.NewProgram(model, tea.WithInput(tty), tea.WithOutput(tty))

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	m, ok := finalModel.(Model)
	if !ok {
		panic("unexpected error")
	}

	// Clear the TUI from the terminal
	viewOutput := m.View()
	lineCount := strings.Count(viewOutput, "\n")
	clearSequence := fmt.Sprintf("\033[%dA\033[J", lineCount)
	_, _ = tty.WriteString(clearSequence)

	// If cancelled (Ctrl-C), exit without saving or outputting
	if m.Cancelled() {
		return "", nil
	}

	// If quitting normally, save entries and output shell commands
	if m.Quitting() {
		// Save is handled by persistEntries in the model
		// Output export/unset commands to stdout
		output := generateShellCommands(m.Entries(), envSnapshot)
		return output, nil
	}

	return "", nil
}

// captureEnvironment builds a snapshot of environment variable values for all
// entry names. Returns a map of name -> value (empty string if not set).
func captureEnvironment(entries []Entry) map[string]string {
	env := make(map[string]string)
	for _, entry := range entries {
		if _, ok := env[entry.Name]; !ok {
			env[entry.Name] = os.Getenv(entry.Name)
		}
	}
	return env
}

// syncWithEnvironment updates the Selected state of each variable based on the
// captured environment snapshot.
//
// A variable is marked as Selected if both its Name and Value match the
// environment. For variables with duplicate names (radio groups), only the
// variable
// whose value matches the environment is selected. If no exact match is found,
// no variable with that name is selected.
func syncWithEnvironment(entries []Entry, env map[string]string) {
	selectedNames := make(map[string]struct{})

	for i := range entries {
		name := entries[i].Name
		envVal := env[name]

		if envVal == "" {
			entries[i].Selected = false
			continue
		}

		// If we already selected a variable for this name, skip
		if _, ok := selectedNames[name]; ok {
			entries[i].Selected = false
			continue
		}

		// Select only if both name and value match the environment
		if entries[i].Value == envVal {
			entries[i].Selected = true
			selectedNames[name] = struct{}{}
		} else {
			entries[i].Selected = false
		}
	}
}

// generateShellCommands produces export and unset statements for the given
// variables. Only outputs commands when the value has actually changed from the
// original environment state.
func generateShellCommands(entries []Entry, env map[string]string) string {
	// Build a map of name -> selected variable (if any) for radio-group
	// handling
	selectedByName := make(map[string]*Entry)
	for i := range entries {
		if entries[i].Selected {
			selectedByName[entries[i].Name] = &entries[i]
		}
	}

	commands := make([]string, 0, len(entries))
	handledNames := make(map[string]struct{})

	for _, entry := range entries {
		if _, ok := handledNames[entry.Name]; ok {
			continue
		}
		handledNames[entry.Name] = struct{}{}

		originalValue := env[entry.Name]

		if selected, ok := selectedByName[entry.Name]; ok {
			// Only export if the value differs from the original
			if selected.Value != originalValue {
				escaped := strings.ReplaceAll(selected.Value, "'", "'\\''")
				commands = append(
					commands,
					fmt.Sprintf("export %s='%s'", selected.Name, escaped),
				)
			}
		} else if originalValue != "" {
			// No variable with this name is selected, unset if it was
			// originally
			// set
			commands = append(commands, fmt.Sprintf("unset %s", entry.Name))
		}
	}

	return strings.Join(commands, "\n")
}

// unlockFile prompts for password or retrieves key from keychain to unlock
// an encrypted file. Returns the encryption key.
func unlockFile(file *entries.File) ([]byte, error) {
	if !file.Encrypted() {
		return nil, fmt.Errorf("file is not encrypted")
	}

	if file.Encryption.Mode == "password" {
		// Check for APIKI_PASSWORD environment variable
		if password := os.Getenv("APIKI_PASSWORD"); password != "" {
			key, err := file.VerifyPassword(password)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid password from APIKI_PASSWORD: %w",
					err,
				)
			}
			return key, nil
		}

		// Prompt for password
		firstAttempt := true
		for {
			password, err := prompt.ReadPassword("Enter password: ")
			if err != nil {
				return nil, fmt.Errorf("failed to read password: %w", err)
			}

			key, err := file.VerifyPassword(password)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Wrong password.\n")
				if firstAttempt {
					firstAttempt = false
					continue
				}

				return nil, fmt.Errorf("too many wrong password attempts")
			}
			return key, nil
		}
	} else if file.Encryption.Mode == "keychain" {
		// Retrieve from keychain (may trigger Touch ID on macOS)
		fmt.Fprintf(os.Stderr, "Unlocking variables with keychain...\n")
		key, err := keychain.Retrieve()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to retrieve key from keychain: %w",
				err,
			)
		}
		return key, nil
	}

	return nil, fmt.Errorf("unknown encryption mode: %q", file.Encryption.Mode)
}
