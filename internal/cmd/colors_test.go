package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
)

func TestColorsCmd_Text(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Netflix",
				Domain: "netflix.com",
				Colors: []api.Color{
					{Hex: "#e50914", Type: "accent", Brightness: 45},
					{Hex: "#000000", Type: "dark", Brightness: 0},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newColorsCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"netflix.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "#e50914") {
		t.Errorf("output missing color hex: %s", output)
	}
}
