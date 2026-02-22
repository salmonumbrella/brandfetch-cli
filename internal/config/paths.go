package config

import (
	"os"
	"path/filepath"
)

const appName = "brandfetch"

// ConfigDir returns the config directory path.
// Uses $XDG_CONFIG_HOME/brandfetch or ~/.config/brandfetch
func ConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, appName), nil
}

// ConfigFilePath returns the path to config.json
func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// EnsureDir creates a directory if it doesn't exist, with mode 0700
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o700)
}
