package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

var (
	logoFormat   string
	logoTheme    string
	logoType     string
	logoFallback string
	logoWidth    int
	logoHeight   int
)

// NewLogoCmd creates the logo command.
func NewLogoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logo <identifier>",
		Short: "Get logo URL for an identifier (uses high-quota Logo API)",
		Long: `Fetch the logo URL for a brand identifier (domain, brand ID, ticker, or ISIN).

This command uses the Logo API which has high quota limits.

Examples:
  brandfetch logo github.com
  brandfetch logo github.com --format png
  brandfetch logo github.com --theme dark
  brandfetch logo id_123 --type icon --format png`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient(clientRequirements{requireClientID: true})
			if err != nil {
				return err
			}
			return runLogoCmd(cmd, args, client)
		},
	}

	addLogoFlags(cmd)
	cmd.AddCommand(newLogoDownloadCmd())

	return cmd
}

// newLogoCmdWithClient creates logo command with injected client (for testing).
func newLogoCmdWithClient(client APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "logo <identifier>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogoCmd(cmd, args, client)
		},
	}
	addLogoFlags(cmd)
	return cmd
}

func runLogoCmd(cmd *cobra.Command, args []string, client APIClient) error {
	identifier := args[0]
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	result, err := client.GetLogo(ctx, api.LogoOptions{
		Identifier: identifier,
		Format:     logoFormat,
		Theme:      logoTheme,
		Type:       logoType,
		Fallback:   logoFallback,
		Width:      logoWidth,
		Height:     logoHeight,
	})
	if err != nil {
		return err
	}

	format, _, err := resolveOutput(cmd)
	if err != nil {
		return err
	}
	logoResult := &output.LogoResult{
		URL:        result.URL,
		Identifier: result.Identifier,
		Format:     result.Format,
		Theme:      result.Theme,
		Type:       result.Type,
		Fallback:   result.Fallback,
		Width:      result.Width,
		Height:     result.Height,
	}

	fmt.Fprint(cmd.OutOrStdout(), output.FormatLogo(logoResult, format))
	if format == output.FormatText {
		fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}

func addLogoFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&logoFormat, "format", "svg", "Logo format: svg, png, webp")
	cmd.Flags().StringVar(&logoTheme, "theme", "light", "Logo theme: light, dark")
	cmd.Flags().StringVar(&logoType, "type", "logo", "Logo type: logo, icon, symbol")
	cmd.Flags().StringVar(&logoFallback, "fallback", "", "Fallback: lettermark, icon, symbol, brandfetch, 404")
	cmd.Flags().IntVar(&logoWidth, "width", 0, "Logo width (px)")
	cmd.Flags().IntVar(&logoHeight, "height", 0, "Logo height (px)")
}
