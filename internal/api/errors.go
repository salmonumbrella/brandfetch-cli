package api

import (
	"errors"
	"fmt"
)

// Sentinel errors for common API error conditions.
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
	ErrRateLimited  = errors.New("rate limited")
)

// APIError represents an API error response.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

// Is implements errors.Is for APIError.
func (e *APIError) Is(target error) bool {
	switch e.StatusCode {
	case 401:
		return errors.Is(target, ErrUnauthorized)
	case 404:
		return errors.Is(target, ErrNotFound)
	case 429:
		return errors.Is(target, ErrRateLimited)
	}
	return false
}

// WrapAPIError creates a user-friendly error from API response.
func WrapAPIError(statusCode int, body string) error {
	switch statusCode {
	case 401:
		return &APIError{
			StatusCode: 401,
			Message:    "Invalid API key. Run `brandfetch auth set` to configure.",
		}
	case 404:
		return &APIError{
			StatusCode: 404,
			Message:    "Brand not found",
		}
	case 429:
		return &APIError{
			StatusCode: 429,
			Message:    "Rate limit exceeded. Try again later.",
		}
	default:
		return &APIError{
			StatusCode: statusCode,
			Message:    body,
		}
	}
}
