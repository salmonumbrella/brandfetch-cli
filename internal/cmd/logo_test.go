package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
)

// MockAPIClient for testing commands
type MockAPIClient struct {
	GetLogoFunc  func(ctx context.Context, domain, format, theme string) (*api.LogoResult, error)
	GetBrandFunc func(ctx context.Context, domain string) (*api.Brand, error)
	SearchFunc   func(ctx context.Context, query string, limit int) ([]api.SearchResult, error)
}

func (m *MockAPIClient) GetLogo(ctx context.Context, domain, format, theme string) (*api.LogoResult, error) {
	return m.GetLogoFunc(ctx, domain, format, theme)
}

func (m *MockAPIClient) GetBrand(ctx context.Context, domain string) (*api.Brand, error) {
	return m.GetBrandFunc(ctx, domain)
}

func (m *MockAPIClient) Search(ctx context.Context, query string, limit int) ([]api.SearchResult, error) {
	return m.SearchFunc(ctx, query, limit)
}

func TestLogoCmd_Text(t *testing.T) {
	mock := &MockAPIClient{
		GetLogoFunc: func(ctx context.Context, domain, format, theme string) (*api.LogoResult, error) {
			return &api.LogoResult{
				URL:    "https://cdn.brandfetch.io/github.com/logo.svg",
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
	if !containsStr(output, "https://cdn.brandfetch.io/github.com/logo.svg") {
		t.Errorf("output missing logo URL: %s", output)
	}
}

func TestLogoCmd_JSON(t *testing.T) {
	mock := &MockAPIClient{
		GetLogoFunc: func(ctx context.Context, domain, format, theme string) (*api.LogoResult, error) {
			return &api.LogoResult{
				URL:    "https://cdn.brandfetch.io/github.com/logo.svg",
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

	if result["url"] != "https://cdn.brandfetch.io/github.com/logo.svg" {
		t.Errorf("JSON url = %v, want expected URL", result["url"])
	}
}
