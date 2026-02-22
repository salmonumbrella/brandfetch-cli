package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	rootCmd := NewRootCmd()
	// Add subcommands like Execute() does
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewLogoCmd())
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "brandfetch") {
		t.Errorf("help output missing 'brandfetch'")
	}
	if !containsStr(output, "--output") {
		t.Errorf("help output missing '--output' flag")
	}
}

func TestRootCmd_OutputFlag(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"--output", "json"})

	// Just check it parses without error
	err := rootCmd.ParseFlags([]string{"--output", "json"})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && strings.Contains(s, substr)
}
