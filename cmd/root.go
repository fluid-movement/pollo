package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pollo",
	Short: "LLM-assisted PO file translation tool",
	Long:  "pollo — a CLI tool for LLM-assisted translation of GNU gettext PO files.",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
}

// writeJSON writes v as JSON to stdout, terminated by a newline.
// HTML escaping is disabled.
func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// writeError writes {"ok": false, "error": msg} to stdout and exits with code 1.
func writeError(msg string) {
	_ = writeJSON(map[string]any{
		"ok":    false,
		"error": msg,
	})
	os.Exit(1)
}

// readFlagOrFile reads a string value from either an inline flag or a file flag.
// Exactly one of val or filePath should be non-empty.
// Returns the string value and validates UTF-8.
func readFlagOrFile(val, filePath, flagName string) (string, error) {
	if val != "" && filePath != "" {
		return "", fmt.Errorf("--%s and --%s-file are mutually exclusive", flagName, flagName)
	}
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("reading --%s-file: %w", flagName, err)
		}
		if !utf8.Valid(data) {
			return "", fmt.Errorf("--%s-file contains invalid UTF-8", flagName)
		}
		return strings.TrimRight(string(data), "\r\n"), nil
	}
	if !utf8.ValidString(val) {
		return "", fmt.Errorf("--%s contains invalid UTF-8", flagName)
	}
	return val, nil
}
