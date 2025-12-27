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
)

// Credentials holds Brandfetch API credentials.
type Credentials struct {
	ClientID string `json:"client_id"` // Logo API key (high quota)
	APIKey   string `json:"api_key"`   // Brand API key (limited quota)
	Source   Source `json:"-"`         // Where credentials were loaded from
}

// ErrNoCredentials is returned when no credentials are found.
var ErrNoCredentials = errors.New("no credentials found: set BRANDFETCH_CLIENT_ID and BRANDFETCH_API_KEY environment variables, or run 'brandfetch auth set'")

// KeychainGetter abstracts keychain access for testing.
type KeychainGetter interface {
	Get(key string) (string, error)
}

// LoadCredentials loads credentials in priority order: env → keychain → file.
func LoadCredentials(keychain KeychainGetter, configFilePath string) (*Credentials, error) {
	// 1. Try environment variables first
	clientID := os.Getenv("BRANDFETCH_CLIENT_ID")
	apiKey := os.Getenv("BRANDFETCH_API_KEY")
	if clientID != "" && apiKey != "" {
		return &Credentials{
			ClientID: clientID,
			APIKey:   apiKey,
			Source:   SourceEnv,
		}, nil
	}

	// 2. Try keychain
	if keychain != nil {
		kcClientID, err1 := keychain.Get("client_id")
		kcAPIKey, err2 := keychain.Get("api_key")
		if err1 == nil && err2 == nil && kcClientID != "" && kcAPIKey != "" {
			return &Credentials{
				ClientID: kcClientID,
				APIKey:   kcAPIKey,
				Source:   SourceKeychain,
			}, nil
		}
	}

	// 3. Try config file
	if configFilePath != "" {
		data, err := os.ReadFile(configFilePath)
		if err == nil {
			var creds Credentials
			if err := json.Unmarshal(data, &creds); err == nil {
				if creds.ClientID != "" && creds.APIKey != "" {
					creds.Source = SourceFile
					return &creds, nil
				}
			}
		}
	}

	return nil, ErrNoCredentials
}

// SaveToFile saves credentials to a JSON file with mode 0600.
func SaveToFile(creds *Credentials, path string) error {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
