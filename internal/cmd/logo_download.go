package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

var (
	logoDownloadPath   string
	logoDownloadDir    string
	logoDownloadSHA256 string
)

// newLogoDownloadCmd creates the logo download subcommand.
func newLogoDownloadCmd() *cobra.Command {
	return newLogoDownloadCmdWithClients(nil, nil)
}

func newLogoDownloadCmdWithClients(client APIClient, httpClient HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <identifier>",
		Short: "Download a logo asset",
		Long: `Download a logo asset using the Logo API CDN.

Examples:
  brandfetch logo download github.com
  brandfetch logo download github.com --format png --path ./logo.png
  brandfetch logo download id_123 --type icon --format png --dir ./assets`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient := client
			if apiClient == nil {
				var err error
				apiClient, err = createClient(clientRequirements{requireClientID: true})
				if err != nil {
					return err
				}
			}
			if httpClient == nil {
				httpClient = http.DefaultClient
			}
			return runLogoDownloadCmd(cmd, args, apiClient, httpClient)
		},
	}

	addLogoFlags(cmd)
	cmd.Flags().StringVar(&logoDownloadPath, "path", "", "Output file path")
	cmd.Flags().StringVar(&logoDownloadDir, "dir", "", "Output directory (defaults to current directory)")
	cmd.Flags().StringVar(&logoDownloadSHA256, "sha256", "", "Verify SHA-256 checksum after download")

	return cmd
}

func runLogoDownloadCmd(cmd *cobra.Command, args []string, client APIClient, httpClient HTTPClient) error {
	identifier := args[0]
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if logoDownloadPath != "" && logoDownloadDir != "" {
		return fmt.Errorf("--path and --dir are mutually exclusive")
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

	path := logoDownloadPath
	if path == "" {
		ext := logoFormat
		if extFromURL := strings.TrimPrefix(getExtensionFromURL(result.URL), "."); extFromURL != "" {
			ext = extFromURL
		}
		if ext == "" {
			ext = "svg"
		}

		filename := sanitizeFileName(identifier)
		path = filename + "." + ext
		if logoDownloadDir != "" {
			path = filepath.Join(logoDownloadDir, path)
		}
	}

	if dir := filepath.Dir(path); dir != "." {
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	err = downloadFile(httpClient, result.URL, path)
	if err != nil {
		return fmt.Errorf("failed to download logo: %w", err)
	}

	if logoDownloadSHA256 != "" {
		var ok bool
		ok, err = verifySHA256(path, logoDownloadSHA256)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("sha256 mismatch for %s", path)
		}
	}

	format, _, err := resolveOutput(cmd)
	if err != nil {
		return err
	}

	if format == output.FormatJSON {
		payload := map[string]string{
			"url":  result.URL,
			"path": path,
		}
		return output.PrintJSON(cmd.OutOrStdout(), payload)
	}

	fmt.Fprintln(cmd.OutOrStdout(), path)
	return nil
}

func sanitizeFileName(value string) string {
	safe := strings.TrimSpace(value)

	// Use filepath.Base first to extract just the filename component
	// This prevents path traversal attacks like "../../../etc/passwd"
	safe = filepath.Base(safe)

	// If Base returns "." or ".." or empty, use default
	if safe == "" || safe == "." || safe == ".." {
		return "logo"
	}

	// Now sanitize the remaining characters for safe filename usage
	safe = strings.ReplaceAll(safe, " ", "-")
	safe = strings.ReplaceAll(safe, ":", "-")

	return safe
}
