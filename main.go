package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/loderunner/apiki/commands/apiki"
	"github.com/loderunner/apiki/commands/decrypt"
	"github.com/loderunner/apiki/commands/encrypt"
	"github.com/loderunner/apiki/commands/rotate"
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
			output, err := apiki.Run(variablesPath)
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
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".apiki", "variables.json"), nil
}
