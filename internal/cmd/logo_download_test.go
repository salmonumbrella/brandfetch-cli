package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
)

func TestLogoDownloadCmd_Text(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "logo-bytes")
	}))
	defer server.Close()

	mock := &MockAPIClient{
		GetLogoFunc: func(ctx context.Context, opts api.LogoOptions) (*api.LogoResult, error) {
			return &api.LogoResult{URL: server.URL + "/logo.svg"}, nil
		},
	}

	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "logo.svg")

	var stdout bytes.Buffer
	cmd := newLogoDownloadCmdWithClients(mock, server.Client())
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"github.com", "--path", outPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(data) != "logo-bytes" {
		t.Errorf("unexpected file contents: %s", string(data))
	}

	if !containsStr(stdout.String(), outPath) {
		t.Errorf("stdout missing output path")
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple domain", "github.com", "github.com"},
		{"with spaces", "my brand", "my-brand"},
		{"with slashes", "foo/bar", "bar"},
		{"with backslashes on unix", "foo\\bar", "foo\\bar"}, // backslash is valid in Unix filenames
		{"with colons", "foo:bar", "foo-bar"},
		{"double dots in name", "foo..bar", "foo..bar"}, // consecutive dots in filename are safe
		{"path traversal attempt", "../../../etc/passwd", "passwd"},
		{"quadruple dots", "....", "...."},      // four dots is a valid filename
		{"empty string", "", "logo"},            // empty becomes default
		{"only double dots", "..", "logo"},      // parent dir reference becomes default
		{"single dot", ".", "logo"},             // current dir reference becomes default
		{"complex traversal", "foo/../bar", "bar"},
		{"deeply nested traversal", "a/b/c/../../../etc/passwd", "passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogoDownloadCmd_SHA256(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "logo-bytes")
	}))
	defer server.Close()

	mock := &MockAPIClient{
		GetLogoFunc: func(ctx context.Context, opts api.LogoOptions) (*api.LogoResult, error) {
			return &api.LogoResult{URL: server.URL + "/logo.svg"}, nil
		},
	}

	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "logo.svg")

	var stdout bytes.Buffer
	cmd := newLogoDownloadCmdWithClients(mock, server.Client())
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"github.com", "--path", outPath, "--sha256", "6ca6e2b588e6eac72bbddfe9a172818a9dce1fe141b5645912838bdec2f9ca98"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}
