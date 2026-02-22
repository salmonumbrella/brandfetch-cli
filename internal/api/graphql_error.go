package api

import (
	"errors"
	"fmt"
	"strings"
)

// ErrGraphQL is the sentinel error for GraphQL errors.
var ErrGraphQL = errors.New("graphql error")

// GraphQLErrorDetail represents a single GraphQL error.
type GraphQLErrorDetail struct {
	Message    string        `json:"message"`
	Path       []interface{} `json:"path,omitempty"`
	Extensions interface{}   `json:"extensions,omitempty"`
}

// GraphQLError represents one or more GraphQL errors.
type GraphQLError struct {
	Errors []GraphQLErrorDetail
}

// Error implements the error interface.
func (e *GraphQLError) Error() string {
	if len(e.Errors) == 0 {
		return "graphql error: unknown"
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("graphql error: %s", e.Errors[0].Message)
	}

	messages := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		messages[i] = err.Message
	}
	return fmt.Sprintf("graphql errors: %s", strings.Join(messages, "; "))
}

// Is implements errors.Is for GraphQLError.
func (e *GraphQLError) Is(target error) bool {
	return target == ErrGraphQL
}

// Unwrap returns the sentinel error.
func (e *GraphQLError) Unwrap() error {
	return ErrGraphQL
}

// NewGraphQLError creates a GraphQLError from raw error maps.
func NewGraphQLError(errs []map[string]interface{}) *GraphQLError {
	details := make([]GraphQLErrorDetail, len(errs))
	for i, e := range errs {
		detail := GraphQLErrorDetail{}
		if msg, ok := e["message"].(string); ok {
			detail.Message = msg
		}
		if path, ok := e["path"].([]interface{}); ok {
			detail.Path = path
		}
		if ext, ok := e["extensions"]; ok {
			detail.Extensions = ext
		}
		details[i] = detail
	}
	return &GraphQLError{Errors: details}
}
