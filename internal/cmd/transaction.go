package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

var transactionCountry string

// NewTransactionCmd creates the transaction command.
func NewTransactionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transaction <label>",
		Short: "Resolve a transaction label to a brand",
		Long: `Match a transaction label to a brand using the Transaction API.

Examples:
  brandfetch transaction "STARBUCKS 1234 SEATTLE WA"
  brandfetch transaction "Spotify USA" --country US`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient(clientRequirements{requireAPIKey: true})
			if err != nil {
				return err
			}
			return runTransactionCmd(cmd, args, client)
		},
	}

	cmd.Flags().StringVar(&transactionCountry, "country", "", "Country code (ISO 3166-1 alpha-2)")

	return cmd
}

func newTransactionCmdWithClient(client APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "transaction <label>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTransactionCmd(cmd, args, client)
		},
	}
	cmd.Flags().StringVar(&transactionCountry, "country", "", "Country code")
	return cmd
}

func runTransactionCmd(cmd *cobra.Command, args []string, client APIClient) error {
	if transactionCountry == "" {
		return fmt.Errorf("--country is required (ISO 3166-1 alpha-2 country code, e.g., US, GB, DE)")
	}

	label := args[0]
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	brand, err := client.CreateTransaction(ctx, label, transactionCountry)
	if err != nil {
		return err
	}

	format, colorize, err := resolveOutput(cmd)
	if err != nil {
		return err
	}
	if format == output.FormatJSON {
		return output.PrintJSON(cmd.OutOrStdout(), brand)
	}

	result := convertBrandToOutput(brand)
	fmt.Fprintln(cmd.OutOrStdout(), output.FormatBrand(result, format, colorize))
	return nil
}
