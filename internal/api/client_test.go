package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetBrand(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/brands/github.com" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test_api_key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":        "GitHub",
			"domain":      "github.com",
			"description": "Code hosting",
			"logos":       []interface{}{},
			"colors":      []interface{}{},
			"fonts":       []interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient("test_client_id", "test_api_key")
	client.baseURL = server.URL

	brand, err := client.GetBrand(context.Background(), "github.com")
	if err != nil {
		t.Fatalf("GetBrand() error = %v", err)
	}

	if brand.Name != "GitHub" {
		t.Errorf("brand.Name = %v, want GitHub", brand.Name)
	}
	if brand.Domain != "github.com" {
		t.Errorf("brand.Domain = %v, want github.com", brand.Domain)
	}
}

func TestClient_GetBrand_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := NewClient("test_client_id", "test_api_key")
	client.baseURL = server.URL

	_, err := client.GetBrand(context.Background(), "nonexistent.com")
	if err == nil {
		t.Fatalf("GetBrand() expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}

func TestClient_GetLogo(t *testing.T) {
	// GetLogo now extracts from brand data, so mock the brand endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test_api_key" {
			t.Errorf("expected api_key auth, got: %s", r.Header.Get("Authorization"))
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":   "GitHub",
			"domain": "github.com",
			"logos": []map[string]interface{}{
				{
					"type":  "logo",
					"theme": "light",
					"formats": []map[string]interface{}{
						{
							"src":    "https://cdn.brandfetch.io/logo.svg",
							"format": "svg",
						},
					},
				},
			},
			"colors": []interface{}{},
			"fonts":  []interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient("test_client_id", "test_api_key")
	client.baseURL = server.URL

	logo, err := client.GetLogo(context.Background(), "github.com", "svg", "light")
	if err != nil {
		t.Fatalf("GetLogo() error = %v", err)
	}

	if logo.URL != "https://cdn.brandfetch.io/logo.svg" {
		t.Errorf("logo.URL = %v, want https://cdn.brandfetch.io/logo.svg", logo.URL)
	}
}

func TestClient_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Search API uses path-based query: /v2/search/{name}?c={clientId}
		if r.URL.Path != "/v2/search/coffee" {
			t.Errorf("path = %v, want /v2/search/coffee", r.URL.Path)
		}
		if r.URL.Query().Get("c") != "test_client_id" {
			t.Errorf("query param c = %v, want test_client_id", r.URL.Query().Get("c"))
		}

		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"name": "Starbucks", "domain": "starbucks.com"},
			{"name": "Dunkin", "domain": "dunkindonuts.com"},
		})
	}))
	defer server.Close()

	client := NewClient("test_client_id", "test_api_key")
	client.baseURL = server.URL

	results, err := client.Search(context.Background(), "coffee", 10)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("len(results) = %d, want 2", len(results))
	}
}

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"github.com", "github.com"},
		{"https://github.com", "github.com"},
		{"http://github.com", "github.com"},
		{"www.github.com", "github.com"},
		{"https://www.github.com/", "github.com"},
		{"GITHUB.COM", "github.com"},
	}

	for _, tt := range tests {
		if got := NormalizeDomain(tt.input); got != tt.want {
			t.Errorf("NormalizeDomain(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
