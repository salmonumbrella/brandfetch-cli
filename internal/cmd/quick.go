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
var quickSHA256 bool
var quickSHA256Manifest string
var quickSHA256ManifestOut string
var quickSHA256ManifestAppend bool
var quickSHA256ManifestVerify bool

// HTTPClient interface for downloading files (allows mocking in tests).
type HTTPClient interface {
	Get(url string) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

// NewQuickCmd creates the quick command.
func NewQuickCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quick <identifier> [identifier...]",
		Short: "Get logos, favicon, colors, and fonts in one command",
		Long: `Fetch the essentials for brand work: SVG logos (light + dark), favicon, colors, and fonts.

This is a convenience command that extracts the most commonly needed brand assets.
Uses the Brand API which has limited quota.

Supports batch mode: pass multiple domains to fetch them all at once.
For text output, each brand is separated by a blank line.
For JSON output, results are returned as an array.
For CSS output, variables are prefixed with brand name.
For Tailwind output, each brand gets a nested object.
For downloads, subdirectories are created per brand.

Examples:
  brandfetch quick stripe.com
  brandfetch quick shopline.com --output json
  brandfetch quick stripe.com --download ./brand-assets/
  brandfetch quick stripe.com --css
  brandfetch quick stripe.com --tailwind
  brandfetch quick stripe.com github.com airbnb.com
  brandfetch quick stripe.com github.com --output json
  brandfetch quick stripe.com github.com --css
  brandfetch quick stripe.com github.com --download ./assets/`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient(clientRequirements{requireAPIKey: true})
			if err != nil {
				return err
			}
			return runQuickCmd(cmd, args, client, http.DefaultClient)
		},
	}

	cmd.Flags().StringVarP(&downloadDir, "download", "d", "", "Download assets to specified directory")
	cmd.Flags().BoolVar(&cssOutput, "css", false, "Output colors and fonts as CSS custom properties")
	cmd.Flags().BoolVar(&tailwindOutput, "tailwind", false, "Output colors and fonts as Tailwind CSS config")
	cmd.Flags().BoolVar(&quickSHA256, "sha256", false, "Write SHA-256 checksum files for downloads")
	cmd.Flags().StringVar(&quickSHA256Manifest, "sha256-manifest", "", "Verify downloads against a SHA-256 manifest file")
	cmd.Flags().StringVar(&quickSHA256ManifestOut, "sha256-manifest-out", "", "Write a SHA-256 manifest file for downloads")
	cmd.Flags().BoolVar(&quickSHA256ManifestAppend, "sha256-manifest-append", false, "Merge checksums into existing manifest")
	cmd.Flags().BoolVar(&quickSHA256ManifestVerify, "sha256-manifest-verify", false, "Fail when checksum verification mismatches")

	return cmd
}

func newQuickCmdWithClient(client APIClient) *cobra.Command {
	return newQuickCmdWithClients(client, http.DefaultClient)
}

func newQuickCmdWithClients(client APIClient, httpClient HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "quick <identifier> [identifier...]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuickCmd(cmd, args, client, httpClient)
		},
	}
	cmd.Flags().StringVarP(&downloadDir, "download", "d", "", "Download assets to specified directory")
	cmd.Flags().BoolVar(&cssOutput, "css", false, "Output colors and fonts as CSS custom properties")
	cmd.Flags().BoolVar(&tailwindOutput, "tailwind", false, "Output colors and fonts as Tailwind CSS config")
	cmd.Flags().BoolVar(&quickSHA256, "sha256", false, "Write SHA-256 checksum files for downloads")
	cmd.Flags().StringVar(&quickSHA256Manifest, "sha256-manifest", "", "Verify downloads against a SHA-256 manifest file")
	cmd.Flags().StringVar(&quickSHA256ManifestOut, "sha256-manifest-out", "", "Write a SHA-256 manifest file for downloads")
	cmd.Flags().BoolVar(&quickSHA256ManifestAppend, "sha256-manifest-append", false, "Merge checksums into existing manifest")
	cmd.Flags().BoolVar(&quickSHA256ManifestVerify, "sha256-manifest-verify", false, "Fail when checksum verification mismatches")
	return cmd
}

func runQuickCmd(cmd *cobra.Command, args []string, client APIClient, httpClient HTTPClient) error {
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

	// Fetch all brands, continuing on error
	var results []*output.QuickResult
	var fetchErrors []string

	for _, domain := range args {
		brand, err := client.GetBrand(ctx, domain)
		if err != nil {
			fetchErrors = append(fetchErrors, fmt.Sprintf("%s: %v", domain, err))
			fmt.Fprintf(cmd.ErrOrStderr(), "Error fetching %s: %v\n", domain, err)
			continue
		}
		results = append(results, convertBrandToQuickResult(brand))
	}

	// If no results, return error summary
	if len(results) == 0 {
		return fmt.Errorf("failed to fetch all domains: %s", strings.Join(fetchErrors, "; "))
	}

	// Output based on format
	format, colorize, err := resolveOutput(cmd)
	if err != nil {
		return err
	}

	if cssOutput {
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatQuickCSSBatch(results))
	} else if tailwindOutput {
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatQuickTailwindBatch(results))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatQuickBatch(results, format, colorize))
	}

	// Download assets if --download flag is specified
	if downloadDir != "" {
		var manifest map[string]string
		var manifestEntries []checksumEntry
		if quickSHA256Manifest != "" {
			var err error
			manifest, err = parseSHA256Manifest(quickSHA256Manifest)
			if err != nil {
				return err
			}
		}
		if err := downloadAssetsBatch(cmd, results, httpClient, manifest, &manifestEntries); err != nil {
			return err
		}
		if quickSHA256ManifestOut != "" {
			if err := writeSHA256Manifest(quickSHA256ManifestOut, manifestEntries, quickSHA256ManifestAppend); err != nil {
				return err
			}
		}
	} else if quickSHA256Manifest != "" {
		return fmt.Errorf("--sha256-manifest requires --download")
	} else if quickSHA256ManifestOut != "" {
		return fmt.Errorf("--sha256-manifest-out requires --download")
	} else if quickSHA256ManifestAppend {
		return fmt.Errorf("--sha256-manifest-append requires --sha256-manifest-out")
	}

	return nil
}

// downloadAssetsBatch downloads logos and favicon for multiple brands to subdirectories.
func downloadAssetsBatch(cmd *cobra.Command, results []*output.QuickResult, httpClient HTTPClient, manifest map[string]string, manifestEntries *[]checksumEntry) error {
	for _, result := range results {
		// For batch mode with multiple results, create subdirectory per brand
		targetDir := downloadDir
		if len(results) > 1 {
			// Use sanitized domain as subdirectory name
			brandDir := sanitizeDirName(result.Domain)
			targetDir = filepath.Join(downloadDir, brandDir)
		}

		if err := downloadAssetsToDir(cmd, result, httpClient, targetDir, manifest, manifestEntries); err != nil {
			return err
		}
	}
	return nil
}

// sanitizeDirName converts a domain to a safe directory name.
func sanitizeDirName(domain string) string {
	// Remove common TLDs and special characters for cleaner directory names
	name := strings.TrimSuffix(domain, ".com")
	name = strings.TrimSuffix(name, ".io")
	name = strings.TrimSuffix(name, ".org")
	name = strings.TrimSuffix(name, ".net")
	name = strings.TrimSuffix(name, ".co")
	// Replace any remaining dots with hyphens
	name = strings.ReplaceAll(name, ".", "-")
	return name
}

// downloadAssetsToDir downloads logos and favicon to the specified directory.
func downloadAssetsToDir(cmd *cobra.Command, result *output.QuickResult, httpClient HTTPClient, targetDir string, manifest map[string]string, manifestEntries *[]checksumEntry) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to create directory %s: %v\n", targetDir, err)
		return err
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
		destPath := filepath.Join(targetDir, d.filename)
		if err := downloadFile(httpClient, d.url, destPath); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to download %s: %v\n", d.filename, err)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "Downloaded: %s\n", destPath)
			if quickSHA256 {
				if err := writeSHA256File(destPath); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to write checksum for %s: %v\n", d.filename, err)
				}
			}
			if manifest != nil {
				if err := verifySHA256ManifestEntry(destPath, downloadDir, manifest); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: checksum verification failed for %s: %v\n", d.filename, err)
					if quickSHA256ManifestVerify {
						return err
					}
				}
			}
			if manifestEntries != nil {
				if entry, err := buildChecksumEntry(destPath, downloadDir); err == nil {
					*manifestEntries = append(*manifestEntries, entry)
				} else {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: failed to compute checksum for %s: %v\n", d.filename, err)
				}
			}
		}
	}
	return nil
}

func writeSHA256File(path string) error {
	sum, err := computeSHA256(path)
	if err != nil {
		return err
	}
	content := fmt.Sprintf("%s  %s\n", sum, filepath.Base(path))
	return os.WriteFile(path+".sha256", []byte(content), 0o644)
}

// downloadFile downloads a file from url and saves it to destPath.
// It sets browser headers to avoid CDN blocks (e.g., CloudFront 403 errors).
func downloadFile(httpClient HTTPClient, fileURL, destPath string) error {
	req, err := http.NewRequest(http.MethodGet, fileURL, nil)
	if err != nil {
		return err
	}

	// Set browser headers to avoid CDN blocks
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/svg+xml,image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := httpClient.Do(req)
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
