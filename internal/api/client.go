package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL     = "https://api.brandfetch.io"
	defaultLogoBaseURL = "https://api.brandfetch.io"
	defaultTimeout     = 30 * time.Second
)

// Client is the Brandfetch API client.
type Client struct {
	clientID    string // Logo API key (high quota)
	apiKey      string // Brand API key (limited quota)
	baseURL     string
	logoBaseURL string
	httpClient  *http.Client
}

// NewClient creates a new Brandfetch API client.
func NewClient(clientID, apiKey string) *Client {
	return &Client{
		clientID:    clientID,
		apiKey:      apiKey,
		baseURL:     defaultBaseURL,
		logoBaseURL: defaultLogoBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Brand represents a brand from the API.
type Brand struct {
	Name        string  `json:"name"`
	Domain      string  `json:"domain"`
	Description string  `json:"description"`
	Claimed     bool    `json:"claimed"`
	Logos       []Logo  `json:"logos"`
	Colors      []Color `json:"colors"`
	Fonts       []Font  `json:"fonts"`
	Links       []Link  `json:"links"`
}

// Logo represents a logo entry.
type Logo struct {
	Type    string       `json:"type"`
	Theme   string       `json:"theme"`
	Formats []LogoFormat `json:"formats"`
}

// LogoFormat represents a specific logo format.
type LogoFormat struct {
	Src        string `json:"src"`
	Background string `json:"background"`
	Format     string `json:"format"`
	Size       int    `json:"size"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
}

// Color represents a brand color.
type Color struct {
	Hex        string `json:"hex"`
	Type       string `json:"type"`
	Brightness int    `json:"brightness"`
}

// Font represents a brand font.
type Font struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Origin string `json:"origin,omitempty"`
}

// Link represents a social/web link.
type Link struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// LogoResult represents a logo API response.
type LogoResult struct {
	URL    string `json:"url"`
	Format string `json:"format"`
	Theme  string `json:"theme"`
	Type   string `json:"type"`
}

// SearchResult represents a search result.
type SearchResult struct {
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Claimed bool   `json:"claimed"`
	Icon    string `json:"icon"`
}

// GetBrand fetches full brand data (uses Brand API).
func (c *Client) GetBrand(ctx context.Context, domain string) (*Brand, error) {
	domain = NormalizeDomain(domain)
	u := fmt.Sprintf("%s/v2/brands/%s", c.baseURL, domain)

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, WrapAPIError(resp.StatusCode, string(body))
	}

	var brand Brand
	if err := json.Unmarshal(body, &brand); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &brand, nil
}

// GetLogo fetches logo URL by extracting from brand data.
// This uses the Brand API which has broader compatibility with API keys.
func (c *Client) GetLogo(ctx context.Context, domain, format, theme string) (*LogoResult, error) {
	// Get full brand data and extract logo
	brand, err := c.GetBrand(ctx, domain)
	if err != nil {
		return nil, err
	}

	// Find the best matching logo
	var bestLogo *LogoFormat
	var bestLogoMeta *Logo

	for i := range brand.Logos {
		logo := &brand.Logos[i]

		// Filter by theme if specified
		if theme != "" && logo.Theme != theme {
			continue
		}

		// Prefer "logo" type over "icon" or "symbol"
		if logo.Type != "logo" && bestLogoMeta != nil && bestLogoMeta.Type == "logo" {
			continue
		}

		for j := range logo.Formats {
			f := &logo.Formats[j]

			// Filter by format if specified
			if format != "" && f.Format != format {
				continue
			}

			// Prefer SVG > PNG > other formats
			if bestLogo == nil {
				bestLogo = f
				bestLogoMeta = logo
			} else if f.Format == "svg" && bestLogo.Format != "svg" {
				bestLogo = f
				bestLogoMeta = logo
			} else if f.Format == "png" && bestLogo.Format != "svg" && bestLogo.Format != "png" {
				bestLogo = f
				bestLogoMeta = logo
			}
		}
	}

	if bestLogo == nil {
		return nil, fmt.Errorf("no logo found for %s", domain)
	}

	return &LogoResult{
		URL:    bestLogo.Src,
		Format: bestLogo.Format,
		Theme:  bestLogoMeta.Theme,
		Type:   bestLogoMeta.Type,
	}, nil
}

// Search searches for brands (uses Search API with clientId auth).
func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// URL encode the query for the path
	encodedQuery := url.PathEscape(query)

	// Search API uses ?c={clientId} for auth
	params := url.Values{}
	params.Set("c", c.clientID)

	u := fmt.Sprintf("%s/v2/search/%s?%s", c.baseURL, encodedQuery, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, WrapAPIError(resp.StatusCode, string(body))
	}

	var results []SearchResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Apply limit client-side if specified
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// NormalizeDomain cleans up a domain string.
func NormalizeDomain(domain string) string {
	domain = strings.ToLower(domain)
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "www.")
	domain = strings.TrimSuffix(domain, "/")
	return domain
}
