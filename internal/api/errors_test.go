package api

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 401,
		Message:    "Unauthorized",
	}

	want := "API error (401): Unauthorized"
	if got := err.Error(); got != want {
		t.Errorf("APIError.Error() = %v, want %v", got, want)
	}
}

func TestAPIError_Is(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "Not found"}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("APIError(404) should match ErrNotFound")
	}

	err401 := &APIError{StatusCode: 401, Message: "Unauthorized"}
	if !errors.Is(err401, ErrUnauthorized) {
		t.Errorf("APIError(401) should match ErrUnauthorized")
	}

	err429 := &APIError{StatusCode: 429, Message: "Rate limited"}
	if !errors.Is(err429, ErrRateLimited) {
		t.Errorf("APIError(429) should match ErrRateLimited")
	}
}

func TestWrapAPIError(t *testing.T) {
	tests := []struct {
		status  int
		body    string
		wantMsg string
	}{
		{401, "bad key", "Invalid API key. Run `brandfetch auth set` to configure."},
		{404, "not found", "Brand not found"},
		{429, "rate limit", "Rate limit exceeded. Try again later."},
		{500, "server error", "API error (500): server error"},
	}

	for _, tt := range tests {
		err := WrapAPIError(tt.status, tt.body)
		if err == nil {
			t.Errorf("WrapAPIError(%d) = nil, want error", tt.status)
			continue
		}
		// Just check it contains expected message
		if got := err.Error(); got != tt.wantMsg && !contains(got, tt.wantMsg) {
			t.Errorf("WrapAPIError(%d).Error() = %v, want contains %v", tt.status, got, tt.wantMsg)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
