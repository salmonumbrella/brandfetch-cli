package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() error = %v", err)
	}

	if !strings.HasSuffix(dir, "brandfetch") {
		t.Errorf("ConfigDir() = %v, want suffix 'brandfetch'", dir)
	}
}

func TestConfigFilePath(t *testing.T) {
	path, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath() error = %v", err)
	}

	if !strings.HasSuffix(path, "config.json") {
		t.Errorf("ConfigFilePath() = %v, want suffix 'config.json'", path)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Use a temp directory for testing
	tmpDir := t.TempDir()
	testConfigDir := filepath.Join(tmpDir, "brandfetch")

	err := EnsureDir(testConfigDir)
	if err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	info, err := os.Stat(testConfigDir)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	if !info.IsDir() {
		t.Errorf("EnsureDir() did not create directory")
	}
}
