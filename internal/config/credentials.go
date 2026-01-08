package config

import (
	"encoding/json"
	"errors"
	"os"
)

// Credential sources
type Source string

const (
	SourceEnv      Source = "env"
	SourceKeychain Source = "keychain"
	SourceFile     Source = "file"
	SourceMixed    Source = "mixed"
)

// Credentials holds Brandfetch API credentials.
type Credentials struct {
	ClientID string `json:"client_id"` // Logo API key (high quota)
	APIKey   string `json:"api_key"`   // Brand API key (limited quota)
	Source   Source `json:"-"`         // Where credentials were loaded from
}

// ErrNoCredentials is returned when no credentials are found.
var ErrNoCredentials = errors.New("no credentials found: set BRANDFETCH_CLIENT_ID/BRANDFETCH_API_KEY or run 'brandfetch auth set'")

// ErrMissingClientID is returned when the Logo API client ID is required but missing.
var ErrMissingClientID = errors.New("missing Brandfetch client ID: set BRANDFETCH_CLIENT_ID or run 'brandfetch auth set'")

// ErrMissingAPIKey is returned when the Brand API key is required but missing.
var ErrMissingAPIKey = errors.New("missing Brandfetch API key: set BRANDFETCH_API_KEY or run 'brandfetch auth set'")

// Requirements indicates which credentials are required for a command.
type Requirements struct {
	RequireClientID bool
	RequireAPIKey   bool
}

// KeychainGetter abstracts keychain access for testing.
type KeychainGetter interface {
	Get(key string) (string, error)
}

// LoadCredentials loads credentials in priority order: env → keychain → file.
// This maintains the previous behavior and requires both keys.
func LoadCredentials(keychain KeychainGetter, configFilePath string) (*Credentials, error) {
	return LoadCredentialsWithOptions(keychain, configFilePath, Requirements{
		RequireClientID: true,
		RequireAPIKey:   true,
	})
}

// LoadCredentialsWithOptions loads credentials and validates against requirements.
func LoadCredentialsWithOptions(keychain KeychainGetter, configFilePath string, req Requirements) (*Credentials, error) {
	var (
		clientID     string
		apiKey       string
		clientSource Source
		apiSource    Source
	)

	// 1. Environment variables
	envClientID := os.Getenv("BRANDFETCH_CLIENT_ID")
	envAPIKey := os.Getenv("BRANDFETCH_API_KEY")
	if envClientID != "" {
		clientID = envClientID
		clientSource = SourceEnv
	}
	if envAPIKey != "" {
		apiKey = envAPIKey
		apiSource = SourceEnv
	}

	// 2. Keychain (if not already set)
	if keychain != nil {
		if clientID == "" {
			if kcClientID, err := keychain.Get("client_id"); err == nil && kcClientID != "" {
				clientID = kcClientID
				clientSource = SourceKeychain
			}
		}
		if apiKey == "" {
			if kcAPIKey, err := keychain.Get("api_key"); err == nil && kcAPIKey != "" {
				apiKey = kcAPIKey
				apiSource = SourceKeychain
			}
		}
	}

	// 3. Config file (if not already set)
	if configFilePath != "" {
		if clientID == "" || apiKey == "" {
			data, err := os.ReadFile(configFilePath)
			if err == nil {
				var fileCreds Credentials
				if err := json.Unmarshal(data, &fileCreds); err == nil {
					if clientID == "" && fileCreds.ClientID != "" {
						clientID = fileCreds.ClientID
						clientSource = SourceFile
					}
					if apiKey == "" && fileCreds.APIKey != "" {
						apiKey = fileCreds.APIKey
						apiSource = SourceFile
					}
				}
			}
		}
	}

	if clientID == "" && apiKey == "" {
		return nil, ErrNoCredentials
	}
	if req.RequireClientID && clientID == "" {
		return nil, ErrMissingClientID
	}
	if req.RequireAPIKey && apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	source := clientSource
	if clientSource == "" {
		source = apiSource
	}
	if clientSource != "" && apiSource != "" && clientSource != apiSource {
		source = SourceMixed
	}

	return &Credentials{
		ClientID: clientID,
		APIKey:   apiKey,
		Source:   source,
	}, nil
}

// SaveToFile saves credentials to a JSON file with mode 0600.
func SaveToFile(creds *Credentials, path string) error {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
