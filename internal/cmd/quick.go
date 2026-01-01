package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

var downloadDir string
var cssOutput bool
var tailwindOutput bool

// HTTPClient interface for downloading files (allows mocking in tests).
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// NewQuickCmd creates the quick command.
func NewQuickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quick <domain>",
		Short: "Get logos, favicon, colors, and fonts in one command",
		Long: `Fetch the essentials for brand work: SVG logos (light + dark), favicon, colors, and fonts.

This is a convenience command that extracts the most commonly needed brand assets.
Uses the Brand API which has limited quota.

Examples:
  brandfetch quick stripe.com
  brandfetch quick shopline.com --output json
  brandfetch quick stripe.com --download ./brand-assets/
  brandfetch quick stripe.com --css
  brandfetch quick stripe.com --tailwind`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient()
			if err != nil {
				return err
			}
			return runQuickCmd(cmd, args, client, http.DefaultClient)
		},
	}

	cmd.Flags().StringVarP(&downloadDir, "download", "d", "", "Download assets to specified directory")
	cmd.Flags().BoolVar(&cssOutput, "css", false, "Output colors and fonts as CSS custom properties")
	cmd.Flags().BoolVar(&tailwindOutput, "tailwind", false, "Output colors and fonts as Tailwind CSS config")

	return cmd
}

func newQuickCmdWithClient(client APIClient) *cobra.Command {
	return newQuickCmdWithClients(client, http.DefaultClient)
}

func newQuickCmdWithClients(client APIClient, httpClient HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "quick <domain>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuickCmd(cmd, args, client, httpClient)
		},
	}
	cmd.Flags().StringVarP(&downloadDir, "download", "d", "", "Download assets to specified directory")
	cmd.Flags().BoolVar(&cssOutput, "css", false, "Output colors and fonts as CSS custom properties")
	cmd.Flags().BoolVar(&tailwindOutput, "tailwind", false, "Output colors and fonts as Tailwind CSS config")
	return cmd
}

func runQuickCmd(cmd *cobra.Command, args []string, client APIClient, httpClient HTTPClient) error {
	domain := args[0]
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Check for mutually exclusive flags
	if cssOutput && outputFormat == "json" {
		return fmt.Errorf("--css and --output json are mutually exclusive")
	}
	if tailwindOutput && outputFormat == "json" {
		return fmt.Errorf("--tailwind and --output json are mutually exclusive")
	}
	if tailwindOutput && cssOutput {
		return fmt.Errorf("--tailwind and --css are mutually exclusive")
	}

	brand, err := client.GetBrand(ctx, domain)
	if err != nil {
		return err
	}

	result := convertBrandToQuickResult(brand)

	// Output based on format
	if cssOutput {
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatQuickCSS(result))
	} else if tailwindOutput {
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatQuickTailwind(result))
	} else {
		format, _ := output.ParseFormat(outputFormat)
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatQuick(result, format))
	}

	// Download assets if --download flag is specified
	if downloadDir != "" {
		downloadAssets(cmd, result, httpClient)
	}

	return nil
}

// downloadAssets downloads logos and favicon to the specified directory.
func downloadAssets(cmd *cobra.Command, result *output.QuickResult, httpClient HTTPClient) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to create directory %s: %v\n", downloadDir, err)
		return
	}

	var downloads []struct {
		url      string
		filename string
	}

	if result.LogoLight != "" {
		downloads = append(downloads, struct {
			url      string
			filename string
		}{result.LogoLight, "logo-light.svg"})
	}

	if result.LogoDark != "" {
		downloads = append(downloads, struct {
			url      string
			filename string
		}{result.LogoDark, "logo-dark.svg"})
	}

	if result.Favicon != "" {
		ext := getExtensionFromURL(result.Favicon)
		downloads = append(downloads, struct {
			url      string
			filename string
		}{result.Favicon, "favicon" + ext})
	}

	for _, d := range downloads {
		destPath := filepath.Join(downloadDir, d.filename)
		if err := downloadFile(httpClient, d.url, destPath); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to download %s: %v\n", d.filename, err)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "Downloaded: %s\n", destPath)
		}
	}
}

// downloadFile downloads a file from url and saves it to destPath.
func downloadFile(httpClient HTTPClient, fileURL, destPath string) error {
	resp, err := httpClient.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// getExtensionFromURL extracts file extension from a URL.
func getExtensionFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	path := parsed.Path
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}

	// Clean up extension (remove query params if any)
	ext = strings.ToLower(ext)
	return ext
}

func convertBrandToQuickResult(brand *api.Brand) *output.QuickResult {
	result := &output.QuickResult{
		Name:   brand.Name,
		Domain: brand.Domain,
	}

	// Find SVG logos for both themes and favicon
	result.LogoLight, result.LogoDark, result.Favicon = findLogos(brand.Logos)

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

	return result
}

// findLogos extracts light logo, dark logo, and favicon URLs from brand logos.
func findLogos(logos []api.Logo) (logoLight, logoDark, favicon string) {
	for _, logo := range logos {
		for _, f := range logo.Formats {
			// Favicon: prefer icon type, any format (usually jpeg/png)
			if logo.Type == "icon" && favicon == "" {
				favicon = f.Src
			}

			// Logos: only SVG format, only "logo" type
			if f.Format != "svg" || logo.Type != "logo" {
				continue
			}

			if logo.Theme == "light" && logoLight == "" {
				logoLight = f.Src
			}
			if logo.Theme == "dark" && logoDark == "" {
				logoDark = f.Src
			}
		}
	}
	return
}
