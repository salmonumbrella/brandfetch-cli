package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

var searchMax int

// NewSearchCmd creates the search command.
func NewSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for brands by name or keyword",
		Long: `Search for brands matching a query string.

Examples:
  brandfetch search coffee
  brandfetch search "tech company" --max 20
  brandfetch search github --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient(clientRequirements{requireClientID: true})
			if err != nil {
				return err
			}
			return runSearchCmd(cmd, args, client)
		},
	}

	cmd.Flags().IntVar(&searchMax, "max", 10, "Maximum number of results")

	return cmd
}

func newSearchCmdWithClient(client APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "search <query>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearchCmd(cmd, args, client)
		},
	}
	cmd.Flags().IntVar(&searchMax, "max", 10, "Maximum number of results")
	return cmd
}

func runSearchCmd(cmd *cobra.Command, args []string, client APIClient) error {
	query := args[0]
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	results, err := client.Search(ctx, query, searchMax)
	if err != nil {
		return err
	}

	format, colorize, err := resolveOutput(cmd)
	if err != nil {
		return err
	}

	// Convert to output types
	var outputResults []output.SearchResult
	for _, r := range results {
		outputResults = append(outputResults, output.SearchResult{
			Name:    r.Name,
			Domain:  r.Domain,
			Icon:    r.Icon,
			Claimed: r.Claimed,
			BrandID: r.BrandID,
		})
	}

	fmt.Fprint(cmd.OutOrStdout(), output.FormatSearch(outputResults, format, colorize))
	return nil
}
