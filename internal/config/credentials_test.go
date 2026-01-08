package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentials_FromEnv(t *testing.T) {
	// Set env vars
	os.Setenv("BRANDFETCH_CLIENT_ID", "env_client_id")
	os.Setenv("BRANDFETCH_API_KEY", "env_api_key")
	defer os.Unsetenv("BRANDFETCH_CLIENT_ID")
	defer os.Unsetenv("BRANDFETCH_API_KEY")

	creds, err := LoadCredentials(nil, "")
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	if creds.ClientID != "env_client_id" {
		t.Errorf("ClientID = %v, want %v", creds.ClientID, "env_client_id")
	}
	if creds.APIKey != "env_api_key" {
		t.Errorf("APIKey = %v, want %v", creds.APIKey, "env_api_key")
	}
	if creds.Source != SourceEnv {
		t.Errorf("Source = %v, want %v", creds.Source, SourceEnv)
	}
}

func TestCredentials_FromFile(t *testing.T) {
	// Clear env vars
	os.Unsetenv("BRANDFETCH_CLIENT_ID")
	os.Unsetenv("BRANDFETCH_API_KEY")

	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	content := `{"client_id": "file_client_id", "api_key": "file_api_key"}`
	err := os.WriteFile(configFile, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	creds, err := LoadCredentials(nil, configFile)
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}

	if creds.ClientID != "file_client_id" {
		t.Errorf("ClientID = %v, want %v", creds.ClientID, "file_client_id")
	}
	if creds.APIKey != "file_api_key" {
		t.Errorf("APIKey = %v, want %v", creds.APIKey, "file_api_key")
	}
	if creds.Source != SourceFile {
		t.Errorf("Source = %v, want %v", creds.Source, SourceFile)
	}
}

func TestCredentials_Missing(t *testing.T) {
	os.Unsetenv("BRANDFETCH_CLIENT_ID")
	os.Unsetenv("BRANDFETCH_API_KEY")

	_, err := LoadCredentials(nil, "/nonexistent/path/config.json")
	if err == nil {
		t.Errorf("LoadCredentials() expected error for missing credentials")
	}
}

func TestCredentials_PartialEnv(t *testing.T) {
	// Only set one env var
	os.Setenv("BRANDFETCH_CLIENT_ID", "partial_client_id")
	os.Unsetenv("BRANDFETCH_API_KEY")
	defer os.Unsetenv("BRANDFETCH_CLIENT_ID")

	creds, err := LoadCredentialsWithOptions(nil, "/nonexistent/path/config.json", Requirements{RequireClientID: true})
	if err != nil {
		t.Fatalf("LoadCredentialsWithOptions() error = %v", err)
	}
	if creds.ClientID != "partial_client_id" {
		t.Errorf("ClientID = %v, want partial_client_id", creds.ClientID)
	}
	if creds.APIKey != "" {
		t.Errorf("APIKey = %v, want empty", creds.APIKey)
	}
}

func TestCredentials_PartialEnv_MissingAPIKey(t *testing.T) {
	os.Setenv("BRANDFETCH_CLIENT_ID", "partial_client_id")
	os.Unsetenv("BRANDFETCH_API_KEY")
	defer os.Unsetenv("BRANDFETCH_CLIENT_ID")

	_, err := LoadCredentialsWithOptions(nil, "/nonexistent/path/config.json", Requirements{RequireAPIKey: true})
	if err == nil {
		t.Errorf("LoadCredentialsWithOptions() expected error for missing API key")
	}
}
