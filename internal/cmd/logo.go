package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
	"github.com/salmonumbrella/brandfetch-cli/internal/config"
	"github.com/salmonumbrella/brandfetch-cli/internal/output"
	"github.com/salmonumbrella/brandfetch-cli/internal/secrets"
)

var (
	logoFormat string
	logoTheme  string
)

// APIClient interface for dependency injection in tests.
type APIClient interface {
	GetLogo(ctx context.Context, domain, format, theme string) (*api.LogoResult, error)
	GetBrand(ctx context.Context, domain string) (*api.Brand, error)
	Search(ctx context.Context, query string, limit int) ([]api.SearchResult, error)
}

// NewLogoCmd creates the logo command.
func NewLogoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logo <domain>",
		Short: "Get logo URL for a domain (uses high-quota Logo API)",
		Long: `Fetch the logo URL for a brand by domain name.

This command uses the Logo API which has high quota limits.

Examples:
  brandfetch logo github.com
  brandfetch logo github.com --format png
  brandfetch logo github.com --theme dark`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}
			return runLogoCmd(cmd, args, client)
		},
	}

	cmd.Flags().StringVar(&logoFormat, "format", "svg", "Logo format: svg, png")
	cmd.Flags().StringVar(&logoTheme, "theme", "light", "Logo theme: light, dark")

	return cmd
}

// newLogoCmdWithClient creates logo command with injected client (for testing).
func newLogoCmdWithClient(client APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "logo <domain>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogoCmd(cmd, args, client)
		},
	}
	cmd.Flags().StringVar(&logoFormat, "format", "svg", "Logo format")
	cmd.Flags().StringVar(&logoTheme, "theme", "light", "Logo theme")
	return cmd
}

func runLogoCmd(cmd *cobra.Command, args []string, client APIClient) error {
	domain := args[0]
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	result, err := client.GetLogo(ctx, domain, logoFormat, logoTheme)
	if err != nil {
		return err
	}

	format, _ := output.ParseFormat(outputFormat)
	logoResult := &output.LogoResult{
		URL:    result.URL,
		Format: result.Format,
		Theme:  result.Theme,
	}

	fmt.Fprint(cmd.OutOrStdout(), output.FormatLogo(logoResult, format))
	if format == output.FormatText {
		fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}

func createClient() (*api.Client, error) {
	// Try to get credentials
	var keychain config.KeychainGetter
	store, err := secrets.NewStore()
	if err == nil {
		keychain = store
	}

	configPath, _ := config.ConfigFilePath()
	creds, err := config.LoadCredentials(keychain, configPath)
	if err != nil {
		return nil, err
	}

	return api.NewClient(creds.ClientID, creds.APIKey), nil
}
