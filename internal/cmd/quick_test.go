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
	DoFunc  func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	return m.GetFunc(url)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	// Fall back to GetFunc for backwards compatibility with existing tests
	return m.GetFunc(req.URL.String())
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

func TestQuickCmd_Download_SHA256(t *testing.T) {
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
	quickSHA256 = true
	quickSHA256Manifest = ""
	defer func() {
		downloadDir = ""
		quickSHA256 = false
		quickSHA256Manifest = ""
	}()

	cmd := newQuickCmdWithClients(mock, mockHTTP)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"stripe.com", "--download", tempDir, "--sha256"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	files := []string{"logo-light.svg", "logo-dark.svg", "favicon.png"}
	for _, f := range files {
		path := filepath.Join(tempDir, f)
		if _, err := os.Stat(path + ".sha256"); os.IsNotExist(err) {
			t.Errorf("expected checksum file for %s", path)
			continue
		}
		sum, err := computeSHA256(path)
		if err != nil {
			t.Fatalf("computeSHA256() error = %v", err)
		}
		data, err := os.ReadFile(path + ".sha256")
		if err != nil {
			t.Fatalf("failed to read checksum file: %v", err)
		}
		if !strings.Contains(string(data), sum) {
			t.Errorf("checksum file missing hash for %s", path)
		}
	}
}

func TestQuickCmd_Download_SHA256Manifest(t *testing.T) {
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
				},
			}, nil
		},
	}

	mockHTTP := &MockHTTPClient{
		GetFunc: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<svg>light logo</svg>")),
			}, nil
		},
	}

	var stdout, stderr bytes.Buffer
	outputFormat = "text"
	downloadDir = tempDir
	manifestPath := filepath.Join(tempDir, "checksums.sha256")
	quickSHA256Manifest = manifestPath
	quickSHA256ManifestOut = ""
	defer func() {
		downloadDir = ""
		quickSHA256Manifest = ""
		quickSHA256ManifestOut = ""
	}()

	cmd := newQuickCmdWithClients(mock, mockHTTP)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"stripe.com", "--download", tempDir, "--sha256-manifest", manifestPath})

	// write manifest after download, but we need it before verification; precompute expected
	sum := "db349b677a1eeaf813d92017e8221a2b39677880af3a8c4d9a12c2ed731531dd"
	manifest := sum + "  logo-light.svg\n"
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if containsStr(stderr.String(), "checksum verification failed") {
		t.Errorf("unexpected checksum failure: %s", stderr.String())
	}
}

func TestQuickCmd_Download_SHA256ManifestAppend(t *testing.T) {
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
				},
			}, nil
		},
	}

	mockHTTP := &MockHTTPClient{
		GetFunc: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<svg>light logo</svg>")),
			}, nil
		},
	}

	var stdout, stderr bytes.Buffer
	outputFormat = "text"
	downloadDir = tempDir
	manifestPath := filepath.Join(tempDir, "checksums.sha256")
	quickSHA256ManifestOut = manifestPath
	quickSHA256ManifestAppend = true
	defer func() {
		downloadDir = ""
		quickSHA256ManifestOut = ""
		quickSHA256ManifestAppend = false
	}()

	existing := "deadbeef  other.svg\n"
	if err := os.WriteFile(manifestPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("failed to seed manifest: %v", err)
	}

	cmd := newQuickCmdWithClients(mock, mockHTTP)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"stripe.com", "--download", tempDir, "--sha256-manifest-out", manifestPath, "--sha256-manifest-append"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}
	if !strings.Contains(string(data), "other.svg") {
		t.Errorf("expected existing entry to remain")
	}
	if !strings.Contains(string(data), "logo-light.svg") {
		t.Errorf("expected new entry to be appended")
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
		name       string
		faviconURL string
		wantExt    string
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

func TestQuickCmd_Tailwind(t *testing.T) {
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
	tailwindOutput = true
	defer func() { tailwindOutput = false }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"stripe.com", "--tailwind"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Check header comments
	if !containsStr(output, "// Tailwind CSS config for Stripe") {
		t.Errorf("output should contain brand name in comment")
	}
	if !containsStr(output, "// Add to your tailwind.config.js theme.extend") {
		t.Errorf("output should contain usage hint comment")
	}

	// Check structure
	if !containsStr(output, "module.exports = {") {
		t.Errorf("output should contain module.exports = {")
	}

	// Check colors section
	if !containsStr(output, "colors: {") {
		t.Errorf("output should contain colors: {")
	}
	if !containsStr(output, "accent: '#635BFF',") {
		t.Errorf("output should contain accent color")
	}
	if !containsStr(output, "dark: '#0A2540',") {
		t.Errorf("output should contain dark color")
	}
	if !containsStr(output, "light: '#FFFFFF',") {
		t.Errorf("output should contain light color")
	}

	// Check fontFamily section
	if !containsStr(output, "fontFamily: {") {
		t.Errorf("output should contain fontFamily: {")
	}
	if !containsStr(output, `title: ['"Sohne Var"', 'sans-serif'],`) {
		t.Errorf("output should contain title font with double quotes and fallback")
	}
	if !containsStr(output, `body: ['"Sohne Var"', 'sans-serif'],`) {
		t.Errorf("output should contain body font with double quotes and fallback")
	}
}

func TestQuickCmd_Tailwind_DuplicateColors(t *testing.T) {
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
	tailwindOutput = true
	defer func() { tailwindOutput = false }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"test.com", "--tailwind"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Duplicate types should use nested object format with all values grouped
	if !containsStr(output, "brand: {") {
		t.Errorf("output should contain brand nested object")
	}
	if !containsStr(output, "1: '#FF0000',") {
		t.Errorf("output should contain 1: '#FF0000'")
	}
	if !containsStr(output, "2: '#00FF00',") {
		t.Errorf("output should contain 2: '#00FF00'")
	}
	if !containsStr(output, "3: '#0000FF',") {
		t.Errorf("output should contain 3: '#0000FF'")
	}

	// Non-duplicate should NOT use nested object
	if !containsStr(output, "light: '#FFFFFF',") {
		t.Errorf("output should contain light color without nesting")
	}
}

func TestQuickCmd_Tailwind_MutuallyExclusiveWithJSON(t *testing.T) {
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
	tailwindOutput = true
	defer func() {
		outputFormat = "text"
		tailwindOutput = false
	}()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"test.com", "--tailwind"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("Execute() should return error for mutually exclusive flags")
	}

	if !containsStr(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention 'mutually exclusive', got: %v", err)
	}
}

func TestQuickCmd_Tailwind_MutuallyExclusiveWithCSS(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return &api.Brand{
				Name:   "Test",
				Domain: "test.com",
			}, nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	tailwindOutput = true
	cssOutput = true
	defer func() {
		tailwindOutput = false
		cssOutput = false
	}()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"test.com", "--tailwind", "--css"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("Execute() should return error for mutually exclusive flags")
	}

	if !containsStr(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention 'mutually exclusive', got: %v", err)
	}
}

func TestQuickCmd_Tailwind_EmptyColorsAndFonts(t *testing.T) {
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
	tailwindOutput = true
	defer func() { tailwindOutput = false }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"minimal.com", "--tailwind"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Should still have valid structure
	if !containsStr(output, "module.exports = {") {
		t.Errorf("output should contain module.exports = {")
	}
	if !containsStr(output, "}") {
		t.Errorf("output should contain closing brace")
	}

	// Should NOT have colors or fontFamily sections
	if containsStr(output, "colors: {") {
		t.Errorf("output should not contain colors section when no colors")
	}
	if containsStr(output, "fontFamily: {") {
		t.Errorf("output should not contain fontFamily section when no fonts")
	}
}

// Batch mode tests

func TestQuickCmd_Batch_Text(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			switch domain {
			case "stripe.com":
				return &api.Brand{
					Name:   "Stripe",
					Domain: "stripe.com",
					Colors: []api.Color{{Hex: "#635BFF", Type: "accent"}},
				}, nil
			case "github.com":
				return &api.Brand{
					Name:   "GitHub",
					Domain: "github.com",
					Colors: []api.Color{{Hex: "#24292f", Type: "dark"}},
				}, nil
			default:
				return nil, errors.New("unknown domain")
			}
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"stripe.com", "github.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Both brands should be present
	if !containsStr(output, "Stripe") {
		t.Errorf("output should contain Stripe")
	}
	if !containsStr(output, "GitHub") {
		t.Errorf("output should contain GitHub")
	}
	if !containsStr(output, "#635BFF") {
		t.Errorf("output should contain Stripe color")
	}
	if !containsStr(output, "#24292f") {
		t.Errorf("output should contain GitHub color")
	}
}

func TestQuickCmd_Batch_JSON(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			switch domain {
			case "stripe.com":
				return &api.Brand{
					Name:   "Stripe",
					Domain: "stripe.com",
					Colors: []api.Color{{Hex: "#635BFF", Type: "accent"}},
				}, nil
			case "github.com":
				return &api.Brand{
					Name:   "GitHub",
					Domain: "github.com",
					Colors: []api.Color{{Hex: "#24292f", Type: "dark"}},
				}, nil
			default:
				return nil, errors.New("unknown domain")
			}
		},
	}

	var stdout bytes.Buffer
	outputFormat = "json"
	defer func() { outputFormat = "text" }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"stripe.com", "github.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should be a JSON array
	var results []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		t.Fatalf("output not valid JSON array: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	if results[0]["name"] != "Stripe" {
		t.Errorf("first result should be Stripe, got %v", results[0]["name"])
	}
	if results[1]["name"] != "GitHub" {
		t.Errorf("second result should be GitHub, got %v", results[1]["name"])
	}
}

func TestQuickCmd_Batch_CSS(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			switch domain {
			case "stripe.com":
				return &api.Brand{
					Name:   "Stripe",
					Domain: "stripe.com",
					Colors: []api.Color{{Hex: "#635BFF", Type: "accent"}},
				}, nil
			case "github.com":
				return &api.Brand{
					Name:   "GitHub",
					Domain: "github.com",
					Colors: []api.Color{{Hex: "#24292f", Type: "dark"}},
				}, nil
			default:
				return nil, errors.New("unknown domain")
			}
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	cssOutput = true
	defer func() { cssOutput = false }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"stripe.com", "github.com", "--css"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Should have brand-prefixed variables
	if !containsStr(output, "--stripe-color-accent: #635BFF;") {
		t.Errorf("output should contain stripe-prefixed color: %s", output)
	}
	if !containsStr(output, "--github-color-dark: #24292f;") {
		t.Errorf("output should contain github-prefixed color: %s", output)
	}
	// Should have brand comments
	if !containsStr(output, "/* Stripe */") {
		t.Errorf("output should contain Stripe comment")
	}
	if !containsStr(output, "/* GitHub */") {
		t.Errorf("output should contain GitHub comment")
	}
}

func TestQuickCmd_Batch_Tailwind(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			switch domain {
			case "stripe.com":
				return &api.Brand{
					Name:   "Stripe",
					Domain: "stripe.com",
					Colors: []api.Color{{Hex: "#635BFF", Type: "accent"}},
				}, nil
			case "github.com":
				return &api.Brand{
					Name:   "GitHub",
					Domain: "github.com",
					Colors: []api.Color{{Hex: "#24292f", Type: "dark"}},
				}, nil
			default:
				return nil, errors.New("unknown domain")
			}
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	tailwindOutput = true
	defer func() { tailwindOutput = false }()

	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"stripe.com", "github.com", "--tailwind"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()

	// Should have nested brand objects
	if !containsStr(output, "stripe: {") {
		t.Errorf("output should contain stripe nested object: %s", output)
	}
	if !containsStr(output, "github: {") {
		t.Errorf("output should contain github nested object: %s", output)
	}
	if !containsStr(output, "accent: '#635BFF',") {
		t.Errorf("output should contain accent color")
	}
	if !containsStr(output, "dark: '#24292f',") {
		t.Errorf("output should contain dark color")
	}
}

func TestQuickCmd_Batch_Download(t *testing.T) {
	tempDir := t.TempDir()

	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			switch domain {
			case "stripe.com":
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
					},
				}, nil
			case "github.com":
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
					},
				}, nil
			default:
				return nil, errors.New("unknown domain")
			}
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
	downloadDir = tempDir
	defer func() { downloadDir = "" }()

	cmd := newQuickCmdWithClients(mock, mockHTTP)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"stripe.com", "github.com", "--download", tempDir})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify subdirectories were created
	stripePath := filepath.Join(tempDir, "stripe", "logo-light.svg")
	if _, err := os.Stat(stripePath); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", stripePath)
	}

	githubPath := filepath.Join(tempDir, "github", "logo-light.svg")
	if _, err := os.Stat(githubPath); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", githubPath)
	}
}

func TestQuickCmd_Batch_SingleDomain_NoSubdirectory(t *testing.T) {
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

	// Single domain should NOT create subdirectory
	directPath := filepath.Join(tempDir, "logo-light.svg")
	if _, err := os.Stat(directPath); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist (no subdirectory for single domain)", directPath)
	}

	// Should NOT have stripe subdirectory
	subDirPath := filepath.Join(tempDir, "stripe")
	if _, err := os.Stat(subDirPath); err == nil {
		t.Errorf("single domain should not create subdirectory")
	}
}

func TestQuickCmd_Batch_PartialFailure(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			switch domain {
			case "stripe.com":
				return &api.Brand{
					Name:   "Stripe",
					Domain: "stripe.com",
					Colors: []api.Color{{Hex: "#635BFF", Type: "accent"}},
				}, nil
			case "invalid.com":
				return nil, errors.New("domain not found")
			default:
				return nil, errors.New("unknown domain")
			}
		},
	}

	var stdout, stderr bytes.Buffer
	outputFormat = "text"
	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"stripe.com", "invalid.com"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() should not fail with partial success: %v", err)
	}

	// Successful result should be in stdout
	output := stdout.String()
	if !containsStr(output, "Stripe") {
		t.Errorf("output should contain successful brand: %s", output)
	}

	// Error should be reported to stderr
	stderrStr := stderr.String()
	if !containsStr(stderrStr, "invalid.com") {
		t.Errorf("stderr should contain failed domain: %s", stderrStr)
	}
}

func TestQuickCmd_Batch_AllFail(t *testing.T) {
	mock := &MockAPIClient{
		GetBrandFunc: func(ctx context.Context, domain string) (*api.Brand, error) {
			return nil, errors.New("domain not found")
		},
	}

	var stdout, stderr bytes.Buffer
	outputFormat = "text"
	cmd := newQuickCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"invalid1.com", "invalid2.com"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("Execute() should fail when all domains fail")
	}

	if !containsStr(err.Error(), "failed to fetch all domains") {
		t.Errorf("error should mention all domains failed: %v", err)
	}
}

func TestSanitizeDirName(t *testing.T) {
	tests := []struct {
		domain string
		want   string
	}{
		{"stripe.com", "stripe"},
		{"github.com", "github"},
		{"example.io", "example"},
		{"test.org", "test"},
		{"api.stripe.com", "api-stripe"},
		{"sub.domain.net", "sub-domain"},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := sanitizeDirName(tt.domain)
			if got != tt.want {
				t.Errorf("sanitizeDirName(%q) = %q, want %q", tt.domain, got, tt.want)
			}
		})
	}
}
