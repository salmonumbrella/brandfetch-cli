package api

import (
	"errors"
	"strings"
	"testing"
)

func TestGraphQLError_Error(t *testing.T) {
	err := &GraphQLError{
		Errors: []GraphQLErrorDetail{
			{Message: "Field not found", Path: []interface{}{"query", "brand"}},
			{Message: "Unauthorized"},
		},
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "Field not found") {
		t.Errorf("error string missing first message: %s", errStr)
	}
	if !strings.Contains(errStr, "Unauthorized") {
		t.Errorf("error string missing second message: %s", errStr)
	}
}

func TestGraphQLError_SingleError(t *testing.T) {
	err := &GraphQLError{
		Errors: []GraphQLErrorDetail{
			{Message: "Not found"},
		},
	}

	errStr := err.Error()
	if errStr != "graphql error: Not found" {
		t.Errorf("unexpected error: %s", errStr)
	}
}

func TestGraphQLError_Is(t *testing.T) {
	err := &GraphQLError{
		Errors: []GraphQLErrorDetail{{Message: "test"}},
	}

	if !errors.Is(err, ErrGraphQL) {
		t.Error("GraphQLError should match ErrGraphQL sentinel")
	}
}

func TestGraphQLError_Unwrap(t *testing.T) {
	err := &GraphQLError{
		Errors: []GraphQLErrorDetail{{Message: "test"}},
	}

	if err.Unwrap() != ErrGraphQL {
		t.Error("Unwrap should return ErrGraphQL")
	}
}

func TestNewGraphQLError(t *testing.T) {
	rawErrors := []map[string]interface{}{
		{"message": "Field error", "path": []interface{}{"query", "brand"}},
		{"message": "Another error"},
	}

	err := NewGraphQLError(rawErrors)
	if len(err.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(err.Errors))
	}
	if err.Errors[0].Message != "Field error" {
		t.Errorf("expected 'Field error', got %s", err.Errors[0].Message)
	}
}
