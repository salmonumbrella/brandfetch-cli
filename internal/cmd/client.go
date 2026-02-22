package cmd

import (
	"context"
	"encoding/json"
	"io"

	"github.com/salmonumbrella/brandfetch-cli/internal/api"
	"github.com/salmonumbrella/brandfetch-cli/internal/config"
	"github.com/salmonumbrella/brandfetch-cli/internal/secrets"
)

// APIClient interface for dependency injection in tests.
type APIClient interface {
	GetLogo(ctx context.Context, opts api.LogoOptions) (*api.LogoResult, error)
	GetBrand(ctx context.Context, identifier string) (*api.Brand, error)
	Search(ctx context.Context, query string, limit int) ([]api.SearchResult, error)
	CreateTransaction(ctx context.Context, label, countryCode string) (*api.Brand, error)
	GraphQL(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error)
	GraphQLRaw(ctx context.Context, body io.Reader) (json.RawMessage, error)
}

type clientRequirements struct {
	requireClientID bool
	requireAPIKey   bool
}

func createClient(req clientRequirements) (*api.Client, error) {
	// Try to get credentials
	var keychain config.KeychainGetter
	store, err := secrets.NewStore()
	if err == nil {
		keychain = store
	}

	configPath, _ := config.ConfigFilePath()
	creds, err := config.LoadCredentialsWithOptions(keychain, configPath, config.Requirements{
		RequireClientID: req.requireClientID,
		RequireAPIKey:   req.requireAPIKey,
	})
	if err != nil {
		return nil, err
	}

	return api.NewClient(creds.ClientID, creds.APIKey), nil
}
