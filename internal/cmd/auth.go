package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/authserver"
	"github.com/salmonumbrella/brandfetch-cli/internal/secrets"
)

var authStdin bool

// SecretsStore interface for dependency injection.
type SecretsStore interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

// NewAuthCmd creates the auth command group.
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage API credentials",
		Long: `Manage Brandfetch API credentials.

Credentials are stored in the OS keychain by default.

Get your API keys at https://brandfetch.com/developers`,
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthSetCmd())
	cmd.AddCommand(newAuthStatusCmd())
	cmd.AddCommand(newAuthClearCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate via browser",
		Long: `Opens a browser window to configure API credentials interactively.

This provides a guided setup experience with:
  - Links to find your API credentials
  - Secure credential storage in keychain

Brandfetch has two API endpoints with separate keys:
  - Logo API: High quota, used for logo and search queries
  - Brand API: Limited quota, used for full brand data (colors, fonts, etc.)

You can configure one or both keys depending on your needs.

Examples:
  brandfetch auth login`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := secrets.NewStore()
			if err != nil {
				return fmt.Errorf("failed to open keychain: %w", err)
			}
			authStdin = false
			return runAuthSetCmd(cmd, store)
		},
	}
}

func newAuthSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set API credentials",
		Long: `Store API credentials in the OS keychain.

You can configure the Logo API client ID, the Brand API key, or both.
For headless environments, use --stdin to read credentials from stdin.

Examples:
  brandfetch auth set          # Opens browser for credential entry
  brandfetch auth set --stdin  # Read from stdin (client_id, then api_key)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := secrets.NewStore()
			if err != nil {
				return fmt.Errorf("failed to open keychain: %w", err)
			}
			return runAuthSetCmd(cmd, store)
		},
	}

	cmd.Flags().BoolVar(&authStdin, "stdin", false, "Read credentials from stdin")

	return cmd
}

func newAuthSetCmdWithStore(store SecretsStore) *cobra.Command {
	cmd := &cobra.Command{
		Use: "set",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthSetCmd(cmd, store)
		},
	}
	cmd.Flags().BoolVar(&authStdin, "stdin", false, "Read from stdin")
	return cmd
}

func runAuthSetCmd(cmd *cobra.Command, store SecretsStore) error {
	var clientID, apiKey string

	if authStdin {
		// Read from stdin
		reader := bufio.NewReader(cmd.InOrStdin())

		line1, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read client_id: %w", err)
		}
		clientID = strings.TrimSpace(line1)

		line2, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read api_key: %w", err)
		}
		apiKey = strings.TrimSpace(line2)
	} else {
		// Browser-based flow
		server, err := authserver.NewServer()
		if err != nil {
			return fmt.Errorf("failed to start auth server: %w", err)
		}
		defer func() { _ = server.Shutdown() }()

		server.Start()
		url := server.URL()

		fmt.Fprintf(cmd.OutOrStdout(), "Opening browser to configure credentials...\n")
		fmt.Fprintf(cmd.OutOrStdout(), "If browser doesn't open, visit: %s\n\n", url)

		// Try to open browser
		openBrowser(url)

		fmt.Fprintf(cmd.OutOrStdout(), "Waiting for credentials...\n")
		creds, err := server.WaitForCredentials(5 * time.Minute)
		if err != nil {
			return err
		}

		clientID = creds.ClientID
		apiKey = creds.APIKey
	}

	if clientID == "" && apiKey == "" {
		return fmt.Errorf("at least one of client_id or api_key is required")
	}

	if clientID != "" {
		if err := store.Set("client_id", clientID); err != nil {
			return fmt.Errorf("failed to store client_id: %w", err)
		}
	}
	if apiKey != "" {
		if err := store.Set("api_key", apiKey); err != nil {
			return fmt.Errorf("failed to store api_key: %w", err)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Credentials saved successfully.")
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show credential status",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := secrets.NewStore()
			if err != nil {
				return fmt.Errorf("failed to open keychain: %w", err)
			}
			return runAuthStatusCmd(cmd, store)
		},
	}
}

func newAuthStatusCmdWithStore(store SecretsStore) *cobra.Command {
	return &cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthStatusCmd(cmd, store)
		},
	}
}

func runAuthStatusCmd(cmd *cobra.Command, store SecretsStore) error {
	clientID, _ := store.Get("client_id")
	apiKey, _ := store.Get("api_key")

	clientStatus := "not configured"
	if clientID != "" {
		clientStatus = "configured"
	}

	apiStatus := "not configured"
	if apiKey != "" {
		apiStatus = "configured"
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Logo API Key (client_id): %s\n", clientStatus)
	fmt.Fprintf(cmd.OutOrStdout(), "Brand API Key (api_key): %s\n", apiStatus)

	return nil
}

func newAuthClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := secrets.NewStore()
			if err != nil {
				return fmt.Errorf("failed to open keychain: %w", err)
			}
			return runAuthClearCmd(cmd, store)
		},
	}
}

func newAuthClearCmdWithStore(store SecretsStore) *cobra.Command {
	return &cobra.Command{
		Use: "clear",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthClearCmd(cmd, store)
		},
	}
}

func runAuthClearCmd(cmd *cobra.Command, store SecretsStore) error {
	_ = store.Delete("client_id")
	_ = store.Delete("api_key")
	fmt.Fprintln(cmd.OutOrStdout(), "Credentials cleared.")
	return nil
}
