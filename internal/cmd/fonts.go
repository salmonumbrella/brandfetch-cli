package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

// NewFontsCmd creates the fonts command.
func NewFontsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fonts <identifier>",
		Short: "Get fonts for an identifier",
		Long: `Fetch the brand fonts for an identifier.

Examples:
  brandfetch fonts github.com
  brandfetch fonts apple.com --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient(clientRequirements{requireAPIKey: true})
			if err != nil {
				return err
			}
			return runFontsCmd(cmd, args, client)
		},
	}
}

func newFontsCmdWithClient(client APIClient) *cobra.Command {
	return &cobra.Command{
		Use:  "fonts <identifier>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFontsCmd(cmd, args, client)
		},
	}
}

func runFontsCmd(cmd *cobra.Command, args []string, client APIClient) error {
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

	var fonts []output.FontInfo
	for _, f := range brand.Fonts {
		fonts = append(fonts, output.FontInfo{
			Name: f.Name,
			Type: f.Type,
		})
	}

	fmt.Fprint(cmd.OutOrStdout(), output.FormatFonts(fonts, format, colorize))
	return nil
}
