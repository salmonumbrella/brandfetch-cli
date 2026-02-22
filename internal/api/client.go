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
	defaultBaseURL        = "https://api.brandfetch.io"
	defaultLogoBaseURL    = "https://cdn.brandfetch.io"
	defaultGraphQLBaseURL = "https://graphql.brandfetch.io/"
	defaultTimeout        = 30 * time.Second
)

// Client is the Brandfetch API client.
type Client struct {
	clientID       string // Logo API key (high quota)
	apiKey         string // Brand API key (limited quota)
	baseURL        string
	logoBaseURL    string
	graphQLBaseURL string
	httpClient     *http.Client
}

// NewClient creates a new Brandfetch API client.
func NewClient(clientID, apiKey string) *Client {
	return &Client{
		clientID:       clientID,
		apiKey:         apiKey,
		baseURL:        defaultBaseURL,
		logoBaseURL:    defaultLogoBaseURL,
		graphQLBaseURL: defaultGraphQLBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Brand represents a brand from the API.
type Brand struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Domain          string                 `json:"domain"`
	Description     string                 `json:"description"`
	LongDescription string                 `json:"longDescription"`
	Claimed         bool                   `json:"claimed"`
	Logos           []Logo                 `json:"logos"`
	Colors          []Color                `json:"colors"`
	Fonts           []Font                 `json:"fonts"`
	Links           []Link                 `json:"links"`
	Images          []Image                `json:"images"`
	Company         map[string]interface{} `json:"company"`
	QualityScore    float64                `json:"qualityScore"`
	IsNSFW          bool                   `json:"isNsfw"`
	URN             string                 `json:"urn"`
}

// Logo represents a logo entry.
type Logo struct {
	Type    string                   `json:"type"`
	Theme   string                   `json:"theme"`
	Formats []LogoFormat             `json:"formats"`
	Tags    []map[string]interface{} `json:"tags,omitempty"`
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
	Name     string `json:"name"`
	Type     string `json:"type"`
	Origin   string `json:"origin,omitempty"`
	OriginID string `json:"originId,omitempty"`
	Weights  []int  `json:"weights,omitempty"`
}

// Link represents a social/web link.
type Link struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Image represents a brand image entry.
type Image struct {
	Type    string                   `json:"type"`
	Formats []LogoFormat             `json:"formats"`
	Tags    []map[string]interface{} `json:"tags,omitempty"`
}

// LogoResult represents a logo API response.
type LogoResult struct {
	URL        string `json:"url"`
	Identifier string `json:"identifier,omitempty"`
	Format     string `json:"format,omitempty"`
	Theme      string `json:"theme,omitempty"`
	Type       string `json:"type,omitempty"`
	Fallback   string `json:"fallback,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
}

// SearchResult represents a search result.
type SearchResult struct {
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Claimed bool   `json:"claimed"`
	Icon    string `json:"icon"`
	BrandID string `json:"brandId"`
}

// LogoOptions controls Logo API URL generation.
type LogoOptions struct {
	Identifier string
	Theme      string
	Type       string
	Fallback   string
	Width      int
	Height     int
	Format     string
}

// GetBrand fetches full brand data (uses Brand API).
func (c *Client) GetBrand(ctx context.Context, domain string) (*Brand, error) {
	identifier := NormalizeIdentifier(domain)
	u := fmt.Sprintf("%s/v2/brands/%s", c.baseURL, url.PathEscape(identifier))

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

// GetLogo returns a Logo API CDN URL based on the provided options.
func (c *Client) GetLogo(ctx context.Context, opts LogoOptions) (*LogoResult, error) {
	_ = ctx
	if strings.TrimSpace(opts.Identifier) == "" {
		return nil, fmt.Errorf("identifier is required")
	}

	url, err := c.BuildLogoURL(opts)
	if err != nil {
		return nil, err
	}

	return &LogoResult{
		URL:        url,
		Identifier: opts.Identifier,
		Format:     opts.Format,
		Theme:      opts.Theme,
		Type:       opts.Type,
		Fallback:   opts.Fallback,
		Width:      opts.Width,
		Height:     opts.Height,
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

// NormalizeIdentifier keeps non-domain identifiers intact while normalizing domains.
func NormalizeIdentifier(identifier string) string {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return trimmed
	}

	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "id_") || strings.HasPrefix(lower, "urn:") {
		return trimmed
	}

	if strings.Contains(lower, "://") || strings.HasPrefix(lower, "www.") {
		return NormalizeDomain(trimmed)
	}

	if strings.Contains(trimmed, ".") {
		return NormalizeDomain(trimmed)
	}

	return trimmed
}

// BuildLogoURL constructs a Logo API CDN URL.
func (c *Client) BuildLogoURL(opts LogoOptions) (string, error) {
	identifier := NormalizeIdentifier(opts.Identifier)
	if identifier == "" {
		return "", fmt.Errorf("identifier is required")
	}
	if strings.TrimSpace(c.clientID) == "" {
		return "", fmt.Errorf("client ID is required for Logo API")
	}

	path := fmt.Sprintf("%s/%s", strings.TrimRight(c.logoBaseURL, "/"), url.PathEscape(identifier))

	segments := []string{}
	if opts.Width > 0 {
		segments = append(segments, "w", fmt.Sprintf("%d", opts.Width))
	}
	if opts.Height > 0 {
		segments = append(segments, "h", fmt.Sprintf("%d", opts.Height))
	}
	if opts.Theme != "" {
		segments = append(segments, "theme", opts.Theme)
	}
	if opts.Fallback != "" {
		segments = append(segments, "fallback", opts.Fallback)
	}

	typeSegment := opts.Type
	if typeSegment == "" && opts.Format != "" {
		typeSegment = "icon"
	}
	if typeSegment != "" {
		if opts.Format != "" {
			typeSegment = fmt.Sprintf("%s.%s", typeSegment, opts.Format)
		}
		segments = append(segments, "type", typeSegment)
	}

	if len(segments) > 0 {
		path = path + "/" + strings.Join(segments, "/")
	}

	params := url.Values{}
	if strings.TrimSpace(c.clientID) != "" {
		params.Set("c", c.clientID)
	}
	if encoded := params.Encode(); encoded != "" {
		path = path + "?" + encoded
	}

	return path, nil
}

// CreateTransaction runs a Transaction API lookup for a merchant label.
func (c *Client) CreateTransaction(ctx context.Context, label, countryCode string) (*Brand, error) {
	payload := map[string]string{
		"transactionLabel": label,
	}
	if countryCode != "" {
		payload["countryCode"] = countryCode
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	u := fmt.Sprintf("%s/v2/brands/transaction", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

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

// GraphQL executes a GraphQL request (used for webhooks).
func (c *Client) GraphQL(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query is required")
	}

	payload := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	u := c.graphQLBaseURL
	req, err := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

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

	var envelope struct {
		Data   json.RawMessage          `json:"data"`
		Errors []map[string]interface{} `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return nil, NewGraphQLError(envelope.Errors)
	}

	return envelope.Data, nil
}

// GraphQLRaw executes a GraphQL request using a raw JSON body stream.
func (c *Client) GraphQLRaw(ctx context.Context, body io.Reader) (json.RawMessage, error) {
	u := c.graphQLBaseURL
	req, err := http.NewRequestWithContext(ctx, "POST", u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, WrapAPIError(resp.StatusCode, string(respBody))
	}

	var envelope struct {
		Data   json.RawMessage          `json:"data"`
		Errors []map[string]interface{} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return nil, NewGraphQLError(envelope.Errors)
	}

	return envelope.Data, nil
}
