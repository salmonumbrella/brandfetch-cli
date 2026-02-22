package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	colorMode    string
)

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brandfetch",
		Short: "Fetch brand assets from the command line",
		Long: `Brandfetch CLI - Fetch logos, colors, and fonts for any company.

Get your API keys at https://brandfetch.com/developers`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", getEnvDefault("BRANDFETCH_OUTPUT", "text"),
		"Output format: text, json")
	cmd.PersistentFlags().StringVar(&colorMode, "color", getEnvDefault("BRANDFETCH_COLOR", "auto"),
		"Color mode: auto, always, never")

	return cmd
}

// Execute runs the root command.
func Execute(args []string) error {
	rootCmd := NewRootCmd()

	// Add subcommands
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewLogoCmd())
	rootCmd.AddCommand(NewBrandCmd())
	rootCmd.AddCommand(NewQuickCmd())
	rootCmd.AddCommand(NewSearchCmd())
	rootCmd.AddCommand(NewColorsCmd())
	rootCmd.AddCommand(NewFontsCmd())
	rootCmd.AddCommand(NewTransactionCmd())
	rootCmd.AddCommand(NewWebhooksCmd())
	rootCmd.AddCommand(NewGraphQLCmd())
	rootCmd.AddCommand(NewAuthCmd())

	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func getEnvDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// GetOutputFormat returns the current output format.
func GetOutputFormat() string {
	return outputFormat
}
