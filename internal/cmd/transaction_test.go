package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
)

func resetTransactionFlags() {
	transactionCountry = ""
}

func TestTransactionCmd_JSON(t *testing.T) {
	resetTransactionFlags()
	mock := &MockAPIClient{
		CreateTransactionFunc: func(ctx context.Context, label, countryCode string) (*api.Brand, error) {
			if label != "SPOTIFY USA" {
				return nil, api.ErrNotFound
			}
			if countryCode != "US" {
				return nil, api.ErrNotFound
			}
			return &api.Brand{Name: "Spotify", Domain: "spotify.com"}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "json"
	defer func() { outputFormat = "text" }()

	cmd := newTransactionCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"SPOTIFY USA", "--country", "US"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if result["name"] != "Spotify" {
		t.Errorf("JSON name = %v, want Spotify", result["name"])
	}
}
