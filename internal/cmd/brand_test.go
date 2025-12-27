package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
)

func TestBrandCmd_Text(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:        "GitHub",
				Domain:      "github.com",
				Description: "Where the world builds software",
				Colors: []api.Color{
					{Hex: "#24292f", Type: "dark"},
				},
				Fonts: []api.Font{
					{Name: "Mona Sans", Type: "title"},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newBrandCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"github.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "GitHub") {
		t.Errorf("output missing brand name")
	}
	if !containsStr(output, "github.com") {
		t.Errorf("output missing domain")
	}
}

func TestBrandCmd_JSON(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "GitHub",
				Domain: "github.com",
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "json"
	defer func() { outputFormat = "text" }()

	cmd := newBrandCmdWithClient(mock)
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

	if result["name"] != "GitHub" {
		t.Errorf("JSON name = %v, want GitHub", result["name"])
	}
}
