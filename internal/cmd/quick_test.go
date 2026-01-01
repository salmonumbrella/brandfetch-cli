package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
)

// MockHTTPClient for testing downloads.
type MockHTTPClient struct {
	GetFunc func(url string) (*http.Response, error)
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	return m.GetFunc(url)
}

func TestQuickCmd_Text(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Stripe",
				Domain: "stripe.com",
				Logos: []api.Logo{
					{
						Type:  "logo",
						Theme: "light",
						Formats: []api.LogoFormat{
							{Src: "https://asset.brandfetch.io/stripe/logo-light.svg", Format: "svg"},
						},
					},
					{
						Type:  "logo",
						Theme: "dark",
						Formats: []api.LogoFormat{
							{Src: "https://asset.brandfetch.io/stripe/logo-dark.svg", Format: "svg"},
						},
					},
					{
						Type:  "icon",
						Theme: "dark",
						Formats: []api.LogoFormat{
							{Src: "https://asset.brandfetch.io/stripe/favicon.png", Format: "png"},
						},
					},
				},
				Colors: []api.Color{
					{Hex: "#635BFF", Type: "accent"},
					{Hex: "#0A2540", Type: "dark"},
				},
				Fonts: []api.Font{
					{Name: "Sohne Var", Type: "title"},
					{Name: "Sohne Var", Type: "body"},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"stripe.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "Stripe") {
		t.Errorf("output missing brand name")
	}
	if !containsStr(output, "logo-light.svg") {
		t.Errorf("output missing light logo URL")
	}
	if !containsStr(output, "logo-dark.svg") {
		t.Errorf("output missing dark logo URL")
	}
	if !containsStr(output, "favicon.png") {
		t.Errorf("output missing favicon URL")
	}
	if !containsStr(output, "#635BFF") {
		t.Errorf("output missing color")
	}
	if !containsStr(output, "Sohne Var") {
		t.Errorf("output missing font")
	}
}

func TestQuickCmd_JSON(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "GitHub",
				Domain: "github.com",
				Logos: []api.Logo{
					{
						Type:  "logo",
						Theme: "light",
						Formats: []api.LogoFormat{
							{Src: "https://asset.brandfetch.io/github/logo-light.svg", Format: "svg"},
						},
					},
					{
						Type:  "logo",
						Theme: "dark",
						Formats: []api.LogoFormat{
							{Src: "https://asset.brandfetch.io/github/logo-dark.svg", Format: "svg"},
						},
					},
				},
				Colors: []api.Color{
					{Hex: "#24292f", Type: "dark", Brightness: 10},
				},
				Fonts: []api.Font{
					{Name: "Mona Sans", Type: "title"},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "json"
	defer func() { outputFormat = "text" }()

	cmd := newQuickCmdWithClient(mock)
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
	if result["logo_light"] != "https://asset.brandfetch.io/github/logo-light.svg" {
		t.Errorf("JSON logo_light = %v, want expected URL", result["logo_light"])
	}
	if result["logo_dark"] != "https://asset.brandfetch.io/github/logo-dark.svg" {
		t.Errorf("JSON logo_dark = %v, want expected URL", result["logo_dark"])
	}
}

func TestQuickCmd_PrefersSVG(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Test",
				Domain: "test.com",
				Logos: []api.Logo{
					{
						Type:  "logo",
						Theme: "light",
						Formats: []api.LogoFormat{
							{Src: "https://example.com/logo.png", Format: "png"},
							{Src: "https://example.com/logo.svg", Format: "svg"},
						},
					},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"test.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "logo.svg") {
		t.Errorf("should include SVG logo: %s", output)
	}
}

func TestQuickCmd_BothThemes(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Test",
				Domain: "test.com",
				Logos: []api.Logo{
					{
						Type:  "logo",
						Theme: "light",
						Formats: []api.LogoFormat{
							{Src: "https://example.com/logo-light.svg", Format: "svg"},
						},
					},
					{
						Type:  "logo",
						Theme: "dark",
						Formats: []api.LogoFormat{
							{Src: "https://example.com/logo-dark.svg", Format: "svg"},
						},
					},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"test.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "logo-light.svg") {
		t.Errorf("should include light theme logo: %s", output)
	}
	if !containsStr(output, "logo-dark.svg") {
		t.Errorf("should include dark theme logo: %s", output)
	}
}

func TestQuickCmd_Favicon(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Test",
				Domain: "test.com",
				Logos: []api.Logo{
					{
						Type:  "icon",
						Theme: "dark",
						Formats: []api.LogoFormat{
							{Src: "https://example.com/favicon.ico", Format: "ico"},
						},
					},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"test.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "Favicon:") {
		t.Errorf("should include Favicon section: %s", output)
	}
	if !containsStr(output, "favicon.ico") {
		t.Errorf("should include favicon URL: %s", output)
	}
}

func TestQuickCmd_Download(t *testing.T) {
	// Create temp directory for downloads
	tempDir := t.TempDir()

	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Stripe",
				Domain: "stripe.com",
				Logos: []api.Logo{
					{
						Type:  "logo",
						Theme: "light",
						Formats: []api.LogoFormat{
							{Src: "https://asset.brandfetch.io/stripe/logo-light.svg", Format: "svg"},
						},
					},
					{
						Type:  "logo",
						Theme: "dark",
						Formats: []api.LogoFormat{
							{Src: "https://asset.brandfetch.io/stripe/logo-dark.svg", Format: "svg"},
						},
					},
					{
						Type:  "icon",
						Theme: "dark",
						Formats: []api.LogoFormat{
							{Src: "https://asset.brandfetch.io/stripe/favicon.png", Format: "png"},
						},
					},
				},
			}, nil
		},
	}

	mockHTTP := &MockHTTPClient{
		GetFunc: func(url string) (*http.Response, error) {
			var content string
			switch {
			case strings.Contains(url, "logo-light"):
				content = "<svg>light logo</svg>"
			case strings.Contains(url, "logo-dark"):
				content = "<svg>dark logo</svg>"
			case strings.Contains(url, "favicon"):
				content = "fake png data"
			default:
				return nil, errors.New("unexpected URL")
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(content)),
			}, nil
		},
	}

	var stdout, stderr bytes.Buffer
	outputFormat = "text"
	downloadDir = tempDir
	defer func() { downloadDir = "" }()

	cmd := newQuickCmdWithClients(mock, mockHTTP)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"stripe.com", "--download", tempDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify files were created
	files := []string{"logo-light.svg", "logo-dark.svg", "favicon.png"}
	for _, f := range files {
		path := filepath.Join(tempDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", path)
		}
	}

	// Verify stderr has download messages
	stderrStr := stderr.String()
	if !containsStr(stderrStr, "Downloaded:") {
		t.Errorf("stderr should contain download messages: %s", stderrStr)
	}

	// Verify stdout still has normal output
	stdoutStr := stdout.String()
	if !containsStr(stdoutStr, "Stripe") {
		t.Errorf("stdout should still contain brand info: %s", stdoutStr)
	}
}

func TestQuickCmd_Download_CreateDir(t *testing.T) {
	// Use a nested directory that doesn't exist
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "brand-assets")

	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Test",
				Domain: "test.com",
				Logos: []api.Logo{
					{
						Type:  "logo",
						Theme: "light",
						Formats: []api.LogoFormat{
							{Src: "https://example.com/logo.svg", Format: "svg"},
						},
					},
				},
			}, nil
		},
	}

	mockHTTP := &MockHTTPClient{
		GetFunc: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<svg>test</svg>")),
			}, nil
		},
	}

	var stdout, stderr bytes.Buffer
	outputFormat = "text"
	downloadDir = nestedDir
	defer func() { downloadDir = "" }()

	cmd := newQuickCmdWithClients(mock, mockHTTP)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"test.com", "--download", nestedDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Errorf("expected directory %s to be created", nestedDir)
	}

	// Verify file was created
	path := filepath.Join(nestedDir, "logo-light.svg")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", path)
	}
}

func TestQuickCmd_Download_Error(t *testing.T) {
	tempDir := t.TempDir()

	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Test",
				Domain: "test.com",
				Logos: []api.Logo{
					{
						Type:  "logo",
						Theme: "light",
						Formats: []api.LogoFormat{
							{Src: "https://example.com/logo.svg", Format: "svg"},
						},
					},
				},
			}, nil
		},
	}

	mockHTTP := &MockHTTPClient{
		GetFunc: func(url string) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}

	var stdout, stderr bytes.Buffer
	outputFormat = "text"
	downloadDir = tempDir
	defer func() { downloadDir = "" }()

	cmd := newQuickCmdWithClients(mock, mockHTTP)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"test.com", "--download", tempDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() should not fail on download error: %v", err)
	}

	// Verify stderr contains error message
	stderrStr := stderr.String()
	if !containsStr(stderrStr, "Error:") {
		t.Errorf("stderr should contain error message: %s", stderrStr)
	}

	// Verify stdout still has normal output (download errors are not fatal)
	stdoutStr := stdout.String()
	if !containsStr(stdoutStr, "Test") {
		t.Errorf("stdout should still contain brand info: %s", stdoutStr)
	}
}

func TestQuickCmd_Download_HTTPError(t *testing.T) {
	tempDir := t.TempDir()

	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Test",
				Domain: "test.com",
				Logos: []api.Logo{
					{
						Type:  "logo",
						Theme: "light",
						Formats: []api.LogoFormat{
							{Src: "https://example.com/logo.svg", Format: "svg"},
						},
					},
				},
			}, nil
		},
	}

	mockHTTP := &MockHTTPClient{
		GetFunc: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader("not found")),
			}, nil
		},
	}

	var stdout, stderr bytes.Buffer
	outputFormat = "text"
	downloadDir = tempDir
	defer func() { downloadDir = "" }()

	cmd := newQuickCmdWithClients(mock, mockHTTP)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"test.com", "--download", tempDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() should not fail on HTTP error: %v", err)
	}

	// Verify stderr contains error message with status code
	stderrStr := stderr.String()
	if !containsStr(stderrStr, "Error:") || !containsStr(stderrStr, "404") {
		t.Errorf("stderr should contain HTTP error: %s", stderrStr)
	}
}

func TestQuickCmd_Download_FaviconExtensions(t *testing.T) {
	tests := []struct {
		name        string
		faviconURL  string
		wantExt     string
	}{
		{"jpeg extension", "https://example.com/favicon.jpeg", "favicon.jpeg"},
		{"jpg extension", "https://example.com/icon.jpg", "favicon.jpg"},
		{"ico extension", "https://example.com/icon.ico", "favicon.ico"},
		{"png extension", "https://example.com/favicon.png", "favicon.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			mock := &MockAPIClient{
				GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
					return &api.Brand{
						Name:   "Test",
						Domain: "test.com",
						Logos: []api.Logo{
							{
								Type:  "icon",
								Theme: "dark",
								Formats: []api.LogoFormat{
									{Src: tt.faviconURL, Format: ""},
								},
							},
						},
					}, nil
				},
			}

			mockHTTP := &MockHTTPClient{
				GetFunc: func(url string) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("fake data")),
					}, nil
				},
			}

			var stdout, stderr bytes.Buffer
			outputFormat = "text"
			downloadDir = tempDir
			defer func() { downloadDir = "" }()

			cmd := newQuickCmdWithClients(mock, mockHTTP)
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs([]string{"test.com", "--download", tempDir})

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			// Verify file was created with correct extension
			path := filepath.Join(tempDir, tt.wantExt)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("expected file %s to exist", path)
			}
		})
	}
}

func TestGetExtensionFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/file.png", ".png"},
		{"https://example.com/file.SVG", ".svg"},
		{"https://example.com/file.jpeg", ".jpeg"},
		{"https://example.com/path/to/file.ico", ".ico"},
		{"https://example.com/file", ""},
		{"https://example.com/file.PNG?query=param", ".png"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := getExtensionFromURL(tt.url)
			if got != tt.want {
				t.Errorf("getExtensionFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestQuickCmd_CSS(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Stripe",
				Domain: "stripe.com",
				Colors: []api.Color{
					{Hex: "#635BFF", Type: "accent"},
					{Hex: "#0A2540", Type: "dark"},
					{Hex: "#FFFFFF", Type: "light"},
				},
				Fonts: []api.Font{
					{Name: "Sohne Var", Type: "title"},
					{Name: "Sohne Var", Type: "body"},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cssOutput = true
	defer func() { cssOutput = false }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"stripe.com", "--css"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Check structure
	if !containsStr(output, ":root {") {
		t.Errorf("output should contain :root { selector")
	}
	if !containsStr(output, "/* Colors */") {
		t.Errorf("output should contain Colors comment")
	}
	if !containsStr(output, "/* Fonts */") {
		t.Errorf("output should contain Fonts comment")
	}

	// Check color variables
	if !containsStr(output, "--color-accent: #635BFF;") {
		t.Errorf("output should contain accent color variable")
	}
	if !containsStr(output, "--color-dark: #0A2540;") {
		t.Errorf("output should contain dark color variable")
	}
	if !containsStr(output, "--color-light: #FFFFFF;") {
		t.Errorf("output should contain light color variable")
	}

	// Check font variables with sans-serif fallback
	if !containsStr(output, "--font-title: 'Sohne Var', sans-serif;") {
		t.Errorf("output should contain title font variable with fallback")
	}
	if !containsStr(output, "--font-body: 'Sohne Var', sans-serif;") {
		t.Errorf("output should contain body font variable with fallback")
	}
}

func TestQuickCmd_CSS_DuplicateColors(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "TestBrand",
				Domain: "test.com",
				Colors: []api.Color{
					{Hex: "#FF0000", Type: "brand"},
					{Hex: "#00FF00", Type: "brand"},
					{Hex: "#0000FF", Type: "brand"},
					{Hex: "#FFFFFF", Type: "light"},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cssOutput = true
	defer func() { cssOutput = false }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"test.com", "--css"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Duplicate types should get numbered
	if !containsStr(output, "--color-brand-1: #FF0000;") {
		t.Errorf("output should contain --color-brand-1")
	}
	if !containsStr(output, "--color-brand-2: #00FF00;") {
		t.Errorf("output should contain --color-brand-2")
	}
	if !containsStr(output, "--color-brand-3: #0000FF;") {
		t.Errorf("output should contain --color-brand-3")
	}

	// Non-duplicate should not have number
	if !containsStr(output, "--color-light: #FFFFFF;") {
		t.Errorf("output should contain --color-light without number")
	}
}

func TestQuickCmd_CSS_DuplicateFonts(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "TestBrand",
				Domain: "test.com",
				Fonts: []api.Font{
					{Name: "Roboto", Type: "body"},
					{Name: "Open Sans", Type: "body"},
				},
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cssOutput = true
	defer func() { cssOutput = false }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"test.com", "--css"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Duplicate font types should get numbered
	if !containsStr(output, "--font-body-1: 'Roboto', sans-serif;") {
		t.Errorf("output should contain --font-body-1")
	}
	if !containsStr(output, "--font-body-2: 'Open Sans', sans-serif;") {
		t.Errorf("output should contain --font-body-2")
	}
}

func TestQuickCmd_CSS_MutuallyExclusiveWithJSON(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Test",
				Domain: "test.com",
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "json"
	cssOutput = true
	defer func() {
		outputFormat = "text"
		cssOutput = false
	}()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"test.com", "--css"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("Execute() should return error for mutually exclusive flags")
	}

	if !containsStr(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention 'mutually exclusive', got: %v", err)
	}
}

func TestQuickCmd_CSS_EmptyColorsAndFonts(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Minimal",
				Domain: "minimal.com",
				// No colors or fonts
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cssOutput = true
	defer func() { cssOutput = false }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"minimal.com", "--css"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Should still have valid CSS structure
	if !containsStr(output, ":root {") {
		t.Errorf("output should contain :root {")
	}
	if !containsStr(output, "}") {
		t.Errorf("output should contain closing brace")
	}

	// Should NOT have comments for empty sections
	if containsStr(output, "/* Colors */") {
		t.Errorf("output should not contain Colors comment when no colors")
	}
	if containsStr(output, "/* Fonts */") {
		t.Errorf("output should not contain Fonts comment when no fonts")
	}
}
