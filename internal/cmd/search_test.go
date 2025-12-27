package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
)

func TestSearchCmd_Text(t *testing.T) {
	mock := &MockAPIClient{
		SearchFunc: func(ctx context.Context, query string, limit int) ([]api.SearchResult, error) {
			return []api.SearchResult{
				{Name: "Starbucks", Domain: "starbucks.com"},
				{Name: "Dunkin", Domain: "dunkindonuts.com"},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newSearchCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"coffee"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "Starbucks") {
		t.Errorf("output missing Starbucks")
	}
	if !containsStr(output, "starbucks.com") {
		t.Errorf("output missing domain")
	}
}

func TestSearchCmd_JSON(t *testing.T) {
	mock := &MockAPIClient{
		SearchFunc: func(ctx context.Context, query string, limit int) ([]api.SearchResult, error) {
			return []api.SearchResult{
				{Name: "Starbucks", Domain: "starbucks.com"},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "json"
	defer func() { outputFormat = "text" }()

	cmd := newSearchCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"coffee"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestSearchCmd_MaxFlag(t *testing.T) {
	var capturedLimit int
	mock := &MockAPIClient{
		SearchFunc: func(ctx context.Context, query string, limit int) ([]api.SearchResult, error) {
			capturedLimit = limit
			return []api.SearchResult{}, nil
		},
	}

	var stdout bytes.Buffer
	cmd := newSearchCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"coffee", "--max", "5"})

	_ = cmd.Execute()

	if capturedLimit != 5 {
		t.Errorf("limit = %d, want 5", capturedLimit)
	}
}
