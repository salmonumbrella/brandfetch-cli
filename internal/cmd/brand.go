package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

// NewBrandCmd creates the brand command.
func NewBrandCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brand <domain>",
		Short: "Get full brand profile for a domain",
		Long: `Fetch comprehensive brand data including logos, colors, fonts, and links.

This command uses the Brand API which has limited quota.

Examples:
  brandfetch brand github.com
  brandfetch brand stripe.com --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}
			return runBrandCmd(cmd, args, client)
		},
	}
	return cmd
}

func newBrandCmdWithClient(client APIClient) *cobra.Command {
	return &cobra.Command{
		Use:  "brand <domain>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBrandCmd(cmd, args, client)
		},
	}
}

func runBrandCmd(cmd *cobra.Command, args []string, client APIClient) error {
	domain := args[0]
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	brand, err := client.GetBrand(ctx, domain)
	if err != nil {
		return err
	}

	format, _ := output.ParseFormat(outputFormat)
	result := convertBrandToOutput(brand)

	fmt.Fprintln(cmd.OutOrStdout(), output.FormatBrand(result, format))
	return nil
}

func convertBrandToOutput(brand *api.Brand) *output.BrandResult {
	result := &output.BrandResult{
		Name:        brand.Name,
		Domain:      brand.Domain,
		Description: brand.Description,
	}

	// Convert logos
	for _, logo := range brand.Logos {
		for _, f := range logo.Formats {
			result.Logos = append(result.Logos, output.LogoInfo{
				Type:   logo.Type,
				Theme:  logo.Theme,
				URL:    f.Src,
				Format: f.Format,
			})
		}
	}

	// Convert colors
	for _, c := range brand.Colors {
		result.Colors = append(result.Colors, output.ColorInfo{
			Hex:        c.Hex,
			Type:       c.Type,
			Brightness: c.Brightness,
		})
	}

	// Convert fonts
	for _, f := range brand.Fonts {
		result.Fonts = append(result.Fonts, output.FontInfo{
			Name: f.Name,
			Type: f.Type,
		})
	}

	// Convert links
	for _, l := range brand.Links {
		result.Links = append(result.Links, output.LinkInfo{
			Name: l.Name,
			URL:  l.URL,
		})
	}

	return result
}
