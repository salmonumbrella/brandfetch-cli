package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

var (
	graphqlQuery     string
	graphqlQueryFile string
	graphqlVariables string
	graphqlStdin     bool
	graphqlStdinRaw  bool
)

// NewGraphQLCmd creates the graphql command.
func NewGraphQLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graphql",
		Short: "Run a GraphQL query against Brandfetch",
		Long: `Run a GraphQL query against the Brandfetch GraphQL API.

Examples:
  brandfetch graphql --query "{ me { id } }"
  brandfetch graphql --query-file ./query.graphql --variables '{"input": {"url": "https://example.com"}}'
  cat query.graphql | brandfetch graphql --stdin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := createClient(clientRequirements{requireAPIKey: true})
			if err != nil {
				return err
			}
			return runGraphQLCmd(cmd, client)
		},
	}

	cmd.Flags().StringVar(&graphqlQuery, "query", "", "GraphQL query string")
	cmd.Flags().StringVar(&graphqlQueryFile, "query-file", "", "Path to GraphQL query file")
	cmd.Flags().StringVar(&graphqlVariables, "variables", "", "JSON variables payload")
	cmd.Flags().BoolVar(&graphqlStdin, "stdin", false, "Read GraphQL query (or JSON payload) from stdin")
	cmd.Flags().BoolVar(&graphqlStdinRaw, "stdin-raw", false, "Stream raw JSON payload from stdin")

	return cmd
}

func newGraphQLCmdWithClient(client APIClient) *cobra.Command {
	cmd := &cobra.Command{
		Use: "graphql",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGraphQLCmd(cmd, client)
		},
	}
	cmd.Flags().StringVar(&graphqlQuery, "query", "", "GraphQL query string")
	cmd.Flags().StringVar(&graphqlQueryFile, "query-file", "", "Path to GraphQL query file")
	cmd.Flags().StringVar(&graphqlVariables, "variables", "", "JSON variables payload")
	cmd.Flags().BoolVar(&graphqlStdin, "stdin", false, "Read GraphQL query (or JSON payload) from stdin")
	cmd.Flags().BoolVar(&graphqlStdinRaw, "stdin-raw", false, "Stream raw JSON payload from stdin")
	return cmd
}

func runGraphQLCmd(cmd *cobra.Command, client APIClient) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if graphqlStdinRaw {
		return runGraphQLRawCmd(cmd, client)
	}

	query, variables, err := resolveGraphQLInput(cmd)
	if err != nil {
		return err
	}

	data, err := client.GraphQL(ctx, query, variables)
	if err != nil {
		return err
	}

	format, colorize, err := resolveOutput(cmd)
	if err != nil {
		return err
	}
	if format == output.FormatText {
		handled, err := printGraphQLText(cmd, data, colorize)
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	}

	var payload interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		_, _ = cmd.OutOrStdout().Write(data)
		fmt.Fprintln(cmd.OutOrStdout())
		return nil
	}

	return output.PrintJSON(cmd.OutOrStdout(), payload)
}

func resolveGraphQLInput(cmd *cobra.Command) (string, map[string]interface{}, error) {
	if graphqlStdinRaw {
		return "", nil, fmt.Errorf("--stdin-raw cannot be combined with --query/--query-file/--variables")
	}
	if graphqlStdin {
		return readGraphQLStdin(cmd.InOrStdin())
	}

	query, err := resolveGraphQLQuery()
	if err != nil {
		return "", nil, err
	}

	var variables map[string]interface{}
	if graphqlVariables != "" {
		if err := json.Unmarshal([]byte(graphqlVariables), &variables); err != nil {
			return "", nil, fmt.Errorf("invalid variables JSON: %w", err)
		}
	}

	return query, variables, nil
}

func resolveGraphQLQuery() (string, error) {
	if graphqlQueryFile != "" {
		data, err := os.ReadFile(graphqlQueryFile)
		if err != nil {
			return "", fmt.Errorf("failed to read query file: %w", err)
		}
		return string(data), nil
	}
	if graphqlQuery == "" {
		return "", fmt.Errorf("--query or --query-file is required")
	}
	return graphqlQuery, nil
}

func readGraphQLStdin(r io.Reader) (string, map[string]interface{}, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	input := strings.TrimSpace(string(data))
	if input == "" {
		return "", nil, fmt.Errorf("stdin is empty")
	}

	if strings.HasPrefix(input, "{") {
		var payload struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables"`
		}
		if err := json.Unmarshal([]byte(input), &payload); err == nil && payload.Query != "" {
			return payload.Query, payload.Variables, nil
		}
	}

	return input, nil, nil
}

func runGraphQLRawCmd(cmd *cobra.Command, client APIClient) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	data, err := client.GraphQLRaw(ctx, cmd.InOrStdin())
	if err != nil {
		return err
	}

	format, colorize, err := resolveOutput(cmd)
	if err != nil {
		return err
	}
	if format == output.FormatText {
		handled, err := printGraphQLText(cmd, data, colorize)
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	}

	var payload interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		_, _ = cmd.OutOrStdout().Write(data)
		fmt.Fprintln(cmd.OutOrStdout())
		return nil
	}

	return output.PrintJSON(cmd.OutOrStdout(), payload)
}

func printGraphQLText(cmd *cobra.Command, data json.RawMessage, colorize bool) (bool, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return false, nil
	}

	if _, ok := root["webhooks"]; ok {
		var result webhookListResponse
		if err := json.Unmarshal(data, &result); err != nil {
			return false, nil
		}
		renderWebhookListText(cmd.OutOrStdout(), result)
		return true, nil
	}

	for key := range root {
		if key == "createWebhook" || key == "addWebhookSubscriptions" || key == "removeWebhookSubscriptions" {
			var result webhookMutationResult
			if err := json.Unmarshal(data, &result); err != nil {
				return false, nil
			}
			renderWebhookMutationText(cmd.OutOrStdout(), key, result)
			return true, nil
		}
	}

	if _, ok := root["brand"]; ok {
		var envelope struct {
			Brand *output.BrandResult `json:"brand"`
		}
		if err := json.Unmarshal(data, &envelope); err == nil && envelope.Brand != nil {
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatBrand(envelope.Brand, output.FormatText, colorize))
			return true, nil
		}
	}

	handledLogos := false
	if logosRaw, ok := root["logos"]; ok {
		var logos []output.LogoInfo
		if err := json.Unmarshal(logosRaw, &logos); err == nil {
			fmt.Fprint(cmd.OutOrStdout(), formatLogosText(logos))
			handledLogos = true
		}
	}
	if linksRaw, ok := root["links"]; ok {
		var links []output.LinkInfo
		if err := json.Unmarshal(linksRaw, &links); err == nil {
			if handledLogos {
				fmt.Fprintln(cmd.OutOrStdout())
			}
			fmt.Fprint(cmd.OutOrStdout(), formatLinksText(links))
			handledLogos = true
		}
	}
	if handledLogos {
		return true, nil
	}

	handled := false
	if colorsRaw, ok := root["colors"]; ok {
		var colors []output.ColorInfo
		if err := json.Unmarshal(colorsRaw, &colors); err == nil {
			fmt.Fprint(cmd.OutOrStdout(), output.FormatColors(colors, output.FormatText, colorize))
			handled = true
		}
	}
	if fontsRaw, ok := root["fonts"]; ok {
		var fonts []output.FontInfo
		if err := json.Unmarshal(fontsRaw, &fonts); err == nil {
			if handled {
				fmt.Fprintln(cmd.OutOrStdout())
			}
			fmt.Fprint(cmd.OutOrStdout(), output.FormatFonts(fonts, output.FormatText, colorize))
			handled = true
		}
	}
	if handled {
		return true, nil
	}

	for _, key := range []string{"brands", "search"} {
		if _, ok := root[key]; ok {
			var envelope map[string][]output.SearchResult
			if err := json.Unmarshal(data, &envelope); err == nil {
				if results, ok := envelope[key]; ok {
					fmt.Fprint(cmd.OutOrStdout(), output.FormatSearch(results, output.FormatText, colorize))
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func formatLogosText(logos []output.LogoInfo) string {
	if len(logos) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Logos:\n")
	for _, l := range logos {
		sb.WriteString(fmt.Sprintf("  - %s (%s): %s\n", l.Type, l.Theme, l.URL))
	}
	return sb.String()
}

func formatLinksText(links []output.LinkInfo) string {
	if len(links) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Links:\n")
	for _, l := range links {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", l.Name, l.URL))
	}
	return sb.String()
}
