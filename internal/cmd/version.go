package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "brandfetch version %s\n", version)
		},
	}
}
