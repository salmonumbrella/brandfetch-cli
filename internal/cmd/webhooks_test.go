package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
	"github.com/spf13/cobra"
)

func resetWebhookFlags() {
	webhookURL = ""
	webhookDescription = ""
	webhookEnabled = true
	webhookEvents = nil
	webhookURN = ""
	webhookSubscriptions = nil
	webhooksListEnabled = false
	webhooksListDisabled = false
	webhooksListEvents = nil
	webhooksListURL = ""
	webhooksListJSONFlat = false
	webhooksListTable = false
	webhooksListTableTruncate = 0
	webhooksListTableColumns = nil
}

func TestWebhooksCreate_Text(t *testing.T) {
	resetWebhookFlags()
	webhookURL = "https://example.com/webhooks"
	webhookEvents = []string{"brand.updated"}
	webhookDescription = "Test webhook"
	webhookEnabled = true

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			if !strings.Contains(query, "createWebhook") {
				return nil, api.ErrNotFound
			}
			input, ok := variables["input"].(map[string]interface{})
			if !ok {
				return nil, api.ErrNotFound
			}
			if input["url"] != "https://example.com/webhooks" {
				return nil, api.ErrNotFound
			}
			data := []byte(`{"createWebhook":{"success":true,"webhook":{"urn":"urn:bf:webhook:123"}}}`)
			return json.RawMessage(data), nil
		},
	}

	outputFormat = "text"
	if err := runWebhooksCreateCmd(cmd, mock); err != nil {
		t.Fatalf("runWebhooksCreateCmd() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "urn:bf:webhook:123") {
		t.Errorf("output missing webhook URN")
	}
}

func TestWebhooksList_Text(t *testing.T) {
	resetWebhookFlags()

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			if !strings.Contains(query, "ListWebhooks") {
				return nil, api.ErrNotFound
			}
			data := []byte(`{"webhooks":{"edges":[{"node":{"urn":"urn:bf:webhook:1","url":"https://example.com/webhooks","enabled":true,"events":["brand.updated"],"description":"Test"}}]}}`)
			return json.RawMessage(data), nil
		},
	}

	outputFormat = "text"
	if err := runWebhooksListCmd(cmd, mock); err != nil {
		t.Fatalf("runWebhooksListCmd() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "urn:bf:webhook:1") {
		t.Errorf("output missing webhook URN")
	}
}

func TestWebhooksList_Filter(t *testing.T) {
	resetWebhookFlags()
	webhooksListEnabled = true
	webhooksListEvents = []string{"brand.updated"}

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"webhooks":{"edges":[` +
				`{"node":{"urn":"urn:bf:webhook:1","url":"https://example.com/1","enabled":true,"events":["brand.updated"],"description":""}},` +
				`{"node":{"urn":"urn:bf:webhook:2","url":"https://example.com/2","enabled":false,"events":["brand.verified"],"description":""}}` +
				`]}}`)
			return json.RawMessage(data), nil
		},
	}

	outputFormat = "text"
	if err := runWebhooksListCmd(cmd, mock); err != nil {
		t.Fatalf("runWebhooksListCmd() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "urn:bf:webhook:1") {
		t.Errorf("output missing enabled webhook")
	}
	if strings.Contains(output, "urn:bf:webhook:2") {
		t.Errorf("output should filter disabled webhook")
	}
}

func TestWebhooksList_FilterURL(t *testing.T) {
	resetWebhookFlags()
	webhooksListURL = "match"

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"webhooks":{"edges":[` +
				`{"node":{"urn":"urn:bf:webhook:1","url":"https://match.example.com","enabled":true,"events":["brand.updated"],"description":""}},` +
				`{"node":{"urn":"urn:bf:webhook:2","url":"https://nope.example.com","enabled":true,"events":["brand.updated"],"description":""}}` +
				`]}}`)
			return json.RawMessage(data), nil
		},
	}

	outputFormat = "text"
	if err := runWebhooksListCmd(cmd, mock); err != nil {
		t.Fatalf("runWebhooksListCmd() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "urn:bf:webhook:1") {
		t.Errorf("output missing url-matched webhook")
	}
	if strings.Contains(output, "urn:bf:webhook:2") {
		t.Errorf("output should filter url-unmatched webhook")
	}
}

func TestWebhooksList_JSONFlat(t *testing.T) {
	resetWebhookFlags()
	webhooksListJSONFlat = true

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"webhooks":{"edges":[{"node":{"urn":"urn:bf:webhook:1","url":"https://example.com","enabled":true,"events":["brand.updated"],"description":""}}]}}`)
			return json.RawMessage(data), nil
		},
	}

	outputFormat = "json"
	if err := runWebhooksListCmd(cmd, mock); err != nil {
		t.Fatalf("runWebhooksListCmd() error = %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed))
	}
	if parsed[0]["urn"] != "urn:bf:webhook:1" {
		t.Errorf("unexpected urn: %v", parsed[0]["urn"])
	}
}

func TestWebhooksList_Table(t *testing.T) {
	resetWebhookFlags()
	webhooksListTable = true

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"webhooks":{"edges":[{"node":{"urn":"urn:bf:webhook:1","url":"https://example.com","enabled":true,"events":["brand.updated"],"description":"Test"}}]}}`)
			return json.RawMessage(data), nil
		},
	}

	outputFormat = "text"
	if err := runWebhooksListCmd(cmd, mock); err != nil {
		t.Fatalf("runWebhooksListCmd() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "URN") || !strings.Contains(output, "URL") {
		t.Errorf("table header missing: %s", output)
	}
	if !strings.Contains(output, "urn:bf:webhook:1") {
		t.Errorf("table row missing webhook")
	}
}

func TestWebhooksList_Table_Truncate(t *testing.T) {
	resetWebhookFlags()
	webhooksListTable = true
	webhooksListTableTruncate = 10

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"webhooks":{"edges":[{"node":{"urn":"urn:bf:webhook:1","url":"https://very-long-example.com/path","enabled":true,"events":["brand.updated","brand.verified"],"description":"A long description"}}]}}`)
			return json.RawMessage(data), nil
		},
	}

	outputFormat = "text"
	if err := runWebhooksListCmd(cmd, mock); err != nil {
		t.Fatalf("runWebhooksListCmd() error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "...") {
		t.Errorf("expected truncated output")
	}
}

func TestWebhooksCreate_InvalidURL(t *testing.T) {
	resetWebhookFlags()
	webhookURL = "not-a-valid-url"
	webhookEvents = []string{"brand.updated"}

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	mock := &MockAPIClient{}

	outputFormat = "text"
	err := runWebhooksCreateCmd(cmd, mock)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "invalid URL") {
		t.Errorf("error should mention invalid URL: %v", err)
	}
}

func TestWebhooksCreate_InvalidURLScheme(t *testing.T) {
	resetWebhookFlags()
	webhookURL = "ftp://example.com/webhooks"
	webhookEvents = []string{"brand.updated"}

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&stdout)

	mock := &MockAPIClient{}

	outputFormat = "text"
	err := runWebhooksCreateCmd(cmd, mock)
	if err == nil {
		t.Fatal("expected error for invalid URL scheme")
	}
	if !strings.Contains(err.Error(), "http") {
		t.Errorf("error should mention http/https: %v", err)
	}
}
