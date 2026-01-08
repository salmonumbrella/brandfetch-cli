package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

// NewColorsCmd creates the colors command.
func NewColorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "colors <identifier>",
		Short: "Get color palette for an identifier",
		Long: `Fetch the brand color palette for an identifier.

Examples:
  brandfetch colors netflix.com
  brandfetch colors stripe.com --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient(clientRequirements{requireAPIKey: true})
			if err != nil {
				return err
			}
			return runColorsCmd(cmd, args, client)
		},
	}
}

func newColorsCmdWithClient(client APIClient) *cobra.Command {
	return &cobra.Command{
		Use:  "colors <identifier>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runColorsCmd(cmd, args, client)
		},
	}
}

func runColorsCmd(cmd *cobra.Command, args []string, client APIClient) error {
	domain := args[0]
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	brand, err := client.GetBrand(ctx, domain)
	if err != nil {
		return err
	}

	format, colorize, err := resolveOutput(cmd)
	if err != nil {
		return err
	}

	var colors []output.ColorInfo
	for _, c := range brand.Colors {
		colors = append(colors, output.ColorInfo{
			Hex:        c.Hex,
			Type:       c.Type,
			Brightness: c.Brightness,
		})
	}

	fmt.Fprint(cmd.OutOrStdout(), output.FormatColors(colors, format, colorize))
	return nil
}
