package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
)

func TestFontsCmd_Text(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "GitHub",
				Domain: "github.com",
				Fonts: []api.Font{
					{Name: "Mona Sans", Type: "title"},
					{Name: "Hubot Sans", Type: "body"},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newFontsCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"github.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "Mona Sans") {
		t.Errorf("output missing font name: %s", output)
	}
}
