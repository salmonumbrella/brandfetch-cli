package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
)

// MockAPIClient for testing commands
type MockAPIClient struct {
	GetLogoFunc           func(ctx context.Context, opts api.LogoOptions) (*api.LogoResult, error)
	GetBrandFunc          func(ctx context.Context, domain string) (*api.Brand, error)
	SearchFunc            func(ctx context.Context, query string, limit int) ([]api.SearchResult, error)
	CreateTransactionFunc func(ctx context.Context, label, countryCode string) (*api.Brand, error)
	GraphQLFunc           func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error)
	GraphQLRawFunc        func(ctx context.Context, body io.Reader) (json.RawMessage, error)
}

func (m *MockAPIClient) GetLogo(ctx context.Context, opts api.LogoOptions) (*api.LogoResult, error) {
	return m.GetLogoFunc(ctx, opts)
}

func (m *MockAPIClient) GetBrand(ctx context.Context, domain string) (*api.Brand, error) {
	return m.GetBrandFunc(ctx, domain)
}

func (m *MockAPIClient) Search(ctx context.Context, query string, limit int) ([]api.SearchResult, error) {
	return m.SearchFunc(ctx, query, limit)
}

func (m *MockAPIClient) CreateTransaction(ctx context.Context, label, countryCode string) (*api.Brand, error) {
	return m.CreateTransactionFunc(ctx, label, countryCode)
}

func (m *MockAPIClient) GraphQL(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
	if m.GraphQLFunc == nil {
		return nil, fmt.Errorf("GraphQL not implemented")
	}
	return m.GraphQLFunc(ctx, query, variables)
}

func (m *MockAPIClient) GraphQLRaw(ctx context.Context, body io.Reader) (json.RawMessage, error) {
	if m.GraphQLRawFunc == nil {
		return nil, fmt.Errorf("GraphQLRaw not implemented")
	}
	return m.GraphQLRawFunc(ctx, body)
}

func TestLogoCmd_Text(t *testing.T) {
	mock := &MockAPIClient{
		GetLogoFunc: func(ctx context.Context, opts api.LogoOptions) (*api.LogoResult, error) {
			if opts.Identifier != "github.com" {
				return nil, fmt.Errorf("unexpected identifier: %s", opts.Identifier)
			}
			if opts.Type != "logo" || opts.Format != "svg" || opts.Theme != "light" {
				return nil, fmt.Errorf("unexpected options")
			}
			return &api.LogoResult{
				URL:    "https://cdn.brandfetch.io/github.com/theme/light/type/logo.svg?c=test_client_id",
				Format: "svg",
				Theme:  "light",
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newLogoCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"github.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "https://cdn.brandfetch.io/github.com/theme/light/type/logo.svg?c=test_client_id") {
		t.Errorf("output missing logo URL: %s", output)
	}
}

func TestLogoCmd_JSON(t *testing.T) {
	mock := &MockAPIClient{
		GetLogoFunc: func(ctx context.Context, opts api.LogoOptions) (*api.LogoResult, error) {
			return &api.LogoResult{
				URL:    "https://cdn.brandfetch.io/github.com/theme/light/type/logo.svg?c=test_client_id",
				Format: "svg",
				Theme:  "light",
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "json" // set global
	defer func() { outputFormat = "text" }()

	cmd := newLogoCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"github.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	if result["url"] != "https://cdn.brandfetch.io/github.com/theme/light/type/logo.svg?c=test_client_id" {
		t.Errorf("JSON url = %v, want expected URL", result["url"])
	}
}
