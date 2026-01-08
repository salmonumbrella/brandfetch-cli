package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/brandfetch-cli/internal/output"
)

var (
	webhookURL                string
	webhookDescription        string
	webhookEnabled            bool
	webhookEvents             []string
	webhookURN                string
	webhookSubscriptions      []string
	webhooksListEnabled       bool
	webhooksListDisabled      bool
	webhooksListEvents        []string
	webhooksListURL           string
	webhooksListJSONFlat      bool
	webhooksListTable         bool
	webhooksListTableTruncate int
	webhooksListTableColumns  []string
)

const createWebhookMutation = `mutation CreateWebhook($input: CreateWebhookInput!) {
  createWebhook(input: $input) {
    code
    message
    success
    webhook {
      urn
      url
      enabled
      events
      description
    }
  }
}`

const addWebhookSubscriptionsMutation = `mutation AddWebhookSubscriptions($input: AddWebhookSubscriptionsInput!) {
  addWebhookSubscriptions(input: $input) {
    code
    message
    success
    webhook {
      urn
    }
  }
}`

const removeWebhookSubscriptionsMutation = `mutation RemoveWebhookSubscriptions($input: RemoveWebhookSubscriptionsInput!) {
  removeWebhookSubscriptions(input: $input) {
    code
    message
    success
    webhook {
      urn
    }
  }
}`

const listWebhooksQuery = `query ListWebhooks {
  webhooks {
    edges {
      node {
        urn
        url
        enabled
        events
        description
      }
    }
  }
}`

// NewWebhooksCmd creates the webhooks command group.
func NewWebhooksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhooks",
		Short: "Manage Brandfetch webhooks (GraphQL)",
		Long: `Create and manage Brandfetch webhooks via the GraphQL API.

Examples:
  brandfetch webhooks create --url https://example.com/webhooks --events brand.updated,brand.verified
  brandfetch webhooks subscribe --webhook urn:bf:webhook:123 --subscriptions urn:bf:brand:abc,urn:bf:brand:def`,
	}

	cmd.AddCommand(newWebhooksCreateCmd())
	cmd.AddCommand(newWebhooksListCmd())
	cmd.AddCommand(newWebhooksSubscribeCmd())
	cmd.AddCommand(newWebhooksUnsubscribeCmd())

	return cmd
}

func newWebhooksCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		RunE: func(cmd *cobra.Command, args []string) error {
			if webhookURL == "" {
				return fmt.Errorf("--url is required")
			}
			if len(webhookEvents) == 0 {
				return fmt.Errorf("--events is required")
			}

			client, err := createClient(clientRequirements{requireAPIKey: true})
			if err != nil {
				return err
			}
			return runWebhooksCreateCmd(cmd, client)
		},
	}

	cmd.Flags().StringVar(&webhookURL, "url", "", "Webhook target URL")
	cmd.Flags().StringSliceVar(&webhookEvents, "events", nil, "Webhook events (comma-separated or repeated)")
	cmd.Flags().StringVar(&webhookDescription, "description", "", "Webhook description")
	cmd.Flags().BoolVar(&webhookEnabled, "enabled", true, "Enable webhook")

	return cmd
}

func newWebhooksSubscribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe a webhook to brand URNs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if webhookURN == "" {
				return fmt.Errorf("--webhook is required")
			}
			if len(webhookSubscriptions) == 0 {
				return fmt.Errorf("--subscriptions is required")
			}

			client, err := createClient(clientRequirements{requireAPIKey: true})
			if err != nil {
				return err
			}
			return runWebhooksSubscribeCmd(cmd, client)
		},
	}

	cmd.Flags().StringVar(&webhookURN, "webhook", "", "Webhook URN")
	cmd.Flags().StringSliceVar(&webhookSubscriptions, "subscriptions", nil, "Brand URNs to subscribe (comma-separated or repeated)")

	return cmd
}

func newWebhooksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if webhooksListEnabled && webhooksListDisabled {
				return fmt.Errorf("--enabled and --disabled are mutually exclusive")
			}
			client, err := createClient(clientRequirements{requireAPIKey: true})
			if err != nil {
				return err
			}
			return runWebhooksListCmd(cmd, client)
		},
	}

	cmd.Flags().BoolVar(&webhooksListEnabled, "enabled", false, "Only show enabled webhooks")
	cmd.Flags().BoolVar(&webhooksListDisabled, "disabled", false, "Only show disabled webhooks")
	cmd.Flags().StringSliceVar(&webhooksListEvents, "event", nil, "Filter by event (repeatable)")
	cmd.Flags().StringVar(&webhooksListURL, "url-contains", "", "Filter by URL substring")
	cmd.Flags().BoolVar(&webhooksListJSONFlat, "json-flat", false, "For JSON output, return a flat array of nodes")
	cmd.Flags().BoolVar(&webhooksListTable, "table", false, "Render output as a table (text only)")
	cmd.Flags().IntVar(&webhooksListTableTruncate, "table-truncate", 0, "Truncate table columns to this width (text only)")
	cmd.Flags().StringSliceVar(&webhooksListTableColumns, "columns", nil, "Table columns (e.g., urn,url,status,events,description)")

	return cmd
}

func newWebhooksUnsubscribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsubscribe",
		Short: "Unsubscribe a webhook from brand URNs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if webhookURN == "" {
				return fmt.Errorf("--webhook is required")
			}
			if len(webhookSubscriptions) == 0 {
				return fmt.Errorf("--subscriptions is required")
			}

			client, err := createClient(clientRequirements{requireAPIKey: true})
			if err != nil {
				return err
			}
			return runWebhooksUnsubscribeCmd(cmd, client)
		},
	}

	cmd.Flags().StringVar(&webhookURN, "webhook", "", "Webhook URN")
	cmd.Flags().StringSliceVar(&webhookSubscriptions, "subscriptions", nil, "Brand URNs to unsubscribe (comma-separated or repeated)")

	return cmd
}

type webhookMutationPayload struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Webhook struct {
		URN string `json:"urn"`
	} `json:"webhook"`
}

type webhookMutationResult struct {
	CreateWebhook              *webhookMutationPayload `json:"createWebhook"`
	AddWebhookSubscriptions    *webhookMutationPayload `json:"addWebhookSubscriptions"`
	RemoveWebhookSubscriptions *webhookMutationPayload `json:"removeWebhookSubscriptions"`
}

type webhookListResponse struct {
	Webhooks struct {
		Edges []struct {
			Node struct {
				URN         string   `json:"urn"`
				URL         string   `json:"url"`
				Enabled     bool     `json:"enabled"`
				Events      []string `json:"events"`
				Description string   `json:"description"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"webhooks"`
}

func runWebhooksCreateCmd(cmd *cobra.Command, client APIClient) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Validate webhook URL
	parsedURL, err := url.Parse(webhookURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid URL: %s (must be a valid HTTP/HTTPS URL)", webhookURL)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s (must be http or https)", parsedURL.Scheme)
	}

	input := map[string]interface{}{
		"url":     webhookURL,
		"events":  normalizeList(webhookEvents),
		"enabled": webhookEnabled,
	}
	if webhookDescription != "" {
		input["description"] = webhookDescription
	}

	data, err := client.GraphQL(ctx, createWebhookMutation, map[string]interface{}{"input": input})
	if err != nil {
		return err
	}

	return renderWebhookResult(cmd, data, "create")
}

func runWebhooksSubscribeCmd(cmd *cobra.Command, client APIClient) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	input := map[string]interface{}{
		"webhookUrn":    webhookURN,
		"subscriptions": normalizeList(webhookSubscriptions),
	}

	data, err := client.GraphQL(ctx, addWebhookSubscriptionsMutation, map[string]interface{}{"input": input})
	if err != nil {
		return err
	}

	return renderWebhookResult(cmd, data, "subscribe")
}

func runWebhooksUnsubscribeCmd(cmd *cobra.Command, client APIClient) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	input := map[string]interface{}{
		"webhookUrn":    webhookURN,
		"subscriptions": normalizeList(webhookSubscriptions),
	}

	data, err := client.GraphQL(ctx, removeWebhookSubscriptionsMutation, map[string]interface{}{"input": input})
	if err != nil {
		return err
	}

	return renderWebhookResult(cmd, data, "unsubscribe")
}

func runWebhooksListCmd(cmd *cobra.Command, client APIClient) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	data, err := client.GraphQL(ctx, listWebhooksQuery, nil)
	if err != nil {
		return err
	}

	format, _, err := resolveOutput(cmd)
	if err != nil {
		return err
	}

	var result webhookListResponse
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	filtered := filterWebhookList(result, webhooksListEnabled, webhooksListDisabled, normalizeList(webhooksListEvents), webhooksListURL)
	if format == output.FormatJSON {
		if webhooksListTable {
			return fmt.Errorf("--table is only supported for text output")
		}
		if webhooksListJSONFlat {
			var flat []interface{}
			for _, edge := range filtered.Webhooks.Edges {
				flat = append(flat, edge.Node)
			}
			return output.PrintJSON(cmd.OutOrStdout(), flat)
		}
		return output.PrintJSON(cmd.OutOrStdout(), filtered)
	}

	if len(filtered.Webhooks.Edges) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No webhooks found.")
		return nil
	}

	if webhooksListTable {
		renderWebhookListTable(cmd.OutOrStdout(), filtered, webhooksListTableTruncate, webhooksListTableColumns)
		return nil
	}

	renderWebhookListText(cmd.OutOrStdout(), filtered)

	return nil
}

func filterWebhookList(result webhookListResponse, enabledOnly, disabledOnly bool, events []string, urlContains string) webhookListResponse {
	if !enabledOnly && !disabledOnly && len(events) == 0 && urlContains == "" {
		return result
	}

	eventSet := make(map[string]struct{}, len(events))
	for _, e := range events {
		if e != "" {
			eventSet[e] = struct{}{}
		}
	}

	var filtered webhookListResponse
	for _, edge := range result.Webhooks.Edges {
		node := edge.Node
		if enabledOnly && !node.Enabled {
			continue
		}
		if disabledOnly && node.Enabled {
			continue
		}
		if urlContains != "" && !strings.Contains(node.URL, urlContains) {
			continue
		}
		if len(eventSet) > 0 {
			match := false
			for _, evt := range node.Events {
				if _, ok := eventSet[evt]; ok {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		filtered.Webhooks.Edges = append(filtered.Webhooks.Edges, edge)
	}

	return filtered
}

func renderWebhookListText(w io.Writer, result webhookListResponse) {
	if len(result.Webhooks.Edges) == 0 {
		fmt.Fprintln(w, "No webhooks found.")
		return
	}

	for _, edge := range result.Webhooks.Edges {
		node := edge.Node
		status := "disabled"
		if node.Enabled {
			status = "enabled"
		}
		events := strings.Join(node.Events, ", ")
		fmt.Fprintf(w, "%s %s (%s)\n", node.URN, node.URL, status)
		if events != "" {
			fmt.Fprintf(w, "  events: %s\n", events)
		}
		if node.Description != "" {
			fmt.Fprintf(w, "  description: %s\n", node.Description)
		}
	}
}

func renderWebhookListTable(w io.Writer, result webhookListResponse, truncateWidth int, columns []string) {
	if len(result.Webhooks.Edges) == 0 {
		fmt.Fprintln(w, "No webhooks found.")
		return
	}

	columns = normalizeColumns(columns)
	headers := []string{"URN", "URL", "STATUS", "EVENTS", "DESCRIPTION"}
	if len(columns) > 0 {
		headers = make([]string, len(columns))
		for i, col := range columns {
			headers[i] = strings.ToUpper(col)
		}
	}
	rows := make([][]string, 0, len(result.Webhooks.Edges))
	for _, edge := range result.Webhooks.Edges {
		node := edge.Node
		status := "disabled"
		if node.Enabled {
			status = "enabled"
		}
		events := strings.Join(node.Events, ", ")
		allCols := map[string]string{
			"urn":         node.URN,
			"url":         node.URL,
			"status":      status,
			"events":      events,
			"description": node.Description,
		}
		if len(columns) == 0 {
			rows = append(rows, []string{node.URN, node.URL, status, events, node.Description})
		} else {
			row := make([]string, len(columns))
			for i, col := range columns {
				row[i] = allCols[col]
			}
			rows = append(rows, row)
		}
	}

	widths := make([]int, len(headers))
	limits := make([]int, len(headers))
	for i := range headers {
		widths[i] = len(headers[i])
		limits[i] = 60
	}
	for i, col := range headers {
		switch strings.ToLower(col) {
		case "status":
			limits[i] = 10
		case "events", "description":
			limits[i] = 40
		}
	}
	if truncateWidth > 0 {
		for i := range limits {
			if limits[i] > truncateWidth {
				limits[i] = truncateWidth
			}
		}
	}
	for _, row := range rows {
		for i, col := range row {
			if len(col) > limits[i] {
				col = truncate(col, limits[i])
			}
			if len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	format := buildTableFormat(widths)
	headerArgs := make([]interface{}, len(headers))
	dividerArgs := make([]interface{}, len(headers))
	for i, header := range headers {
		headerArgs[i] = header
		dividerArgs[i] = strings.Repeat("-", widths[i])
	}
	fmt.Fprintf(w, format, headerArgs...)
	fmt.Fprintf(w, format, dividerArgs...)

	for _, row := range rows {
		cols := make([]string, len(row))
		for i, col := range row {
			cols[i] = truncate(col, limits[i])
		}
		args := make([]interface{}, len(cols))
		for i, col := range cols {
			args[i] = col
		}
		fmt.Fprintf(w, format, args...)
	}
}

func truncate(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func normalizeColumns(columns []string) []string {
	if len(columns) == 0 {
		return nil
	}
	valid := map[string]bool{
		"urn":         true,
		"url":         true,
		"status":      true,
		"events":      true,
		"description": true,
	}
	var normalized []string
	for _, col := range columns {
		for _, part := range strings.Split(col, ",") {
			name := strings.ToLower(strings.TrimSpace(part))
			if name == "" || !valid[name] {
				continue
			}
			normalized = append(normalized, name)
		}
	}
	return normalized
}

func buildTableFormat(widths []int) string {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = fmt.Sprintf("%%-%ds", width)
	}
	return strings.Join(parts, "  ") + "\n"
}

func renderWebhookMutationText(w io.Writer, key string, result webhookMutationResult) {
	var payload *webhookMutationPayload
	action := "webhook"
	switch key {
	case "createWebhook":
		payload = result.CreateWebhook
		action = "create"
	case "addWebhookSubscriptions":
		payload = result.AddWebhookSubscriptions
		action = "subscribe"
	case "removeWebhookSubscriptions":
		payload = result.RemoveWebhookSubscriptions
		action = "unsubscribe"
	}

	if payload == nil {
		fmt.Fprintf(w, "Webhook %s failed: unknown response\n", action)
		return
	}
	if payload.Success {
		fmt.Fprintf(w, "Webhook %s successful: %s\n", action, payload.Webhook.URN)
		return
	}
	message := strings.TrimSpace(payload.Message)
	if message == "" {
		message = "unknown error"
	}
	fmt.Fprintf(w, "Webhook %s failed: %s\n", action, message)
}

func renderWebhookResult(cmd *cobra.Command, data json.RawMessage, action string) error {
	format, _, err := resolveOutput(cmd)
	if err != nil {
		return err
	}
	if format == output.FormatJSON {
		var payload interface{}
		if err := json.Unmarshal(data, &payload); err == nil {
			return output.PrintJSON(cmd.OutOrStdout(), payload)
		}
		_, _ = cmd.OutOrStdout().Write(data)
		fmt.Fprintln(cmd.OutOrStdout())
		return nil
	}

	var result webhookMutationResult
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	payload := result.CreateWebhook
	if action == "subscribe" {
		payload = result.AddWebhookSubscriptions
	}
	if action == "unsubscribe" {
		payload = result.RemoveWebhookSubscriptions
	}
	if payload == nil {
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	if payload.Success {
		fmt.Fprintf(cmd.OutOrStdout(), "Webhook %s successful: %s\n", action, payload.Webhook.URN)
		return nil
	}

	message := strings.TrimSpace(payload.Message)
	if message == "" {
		message = "unknown error"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Webhook %s failed: %s\n", action, message)
	return nil
}

func normalizeList(values []string) []string {
	var out []string
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}
