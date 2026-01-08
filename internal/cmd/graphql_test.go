package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestGraphQLCmd_JSON(t *testing.T) {
	resetGraphQLFlags()
	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			if !strings.Contains(query, "query Test") {
				return nil, errors.New("unexpected query")
			}
			data := []byte(`{"viewer":{"id":"user_123"}}`)
			return json.RawMessage(data), nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "json"
	defer func() { outputFormat = "text" }()

	cmd := newGraphQLCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--query", "query Test { viewer { id } }"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	viewer, _ := parsed["viewer"].(map[string]interface{})
	if viewer["id"] != "user_123" {
		t.Errorf("viewer.id = %v, want user_123", viewer["id"])
	}
}

func TestGraphQLCmd_Text_WebhooksList(t *testing.T) {
	resetGraphQLFlags()
	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"webhooks":{"edges":[{"node":{"urn":"urn:bf:webhook:1","url":"https://example.com","enabled":true,"events":["brand.updated"],"description":"Test"}}]}}`)
			return json.RawMessage(data), nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	defer func() { outputFormat = "text" }()

	cmd := newGraphQLCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--query", "query { webhooks { edges { node { urn url enabled events description } } } }"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "urn:bf:webhook:1") {
		t.Errorf("stdout missing webhook URN")
	}
}

func TestGraphQLCmd_Text_Brand(t *testing.T) {
	resetGraphQLFlags()
	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"brand":{"name":"GitHub","domain":"github.com","description":"Code hosting"}}`)
			return json.RawMessage(data), nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	defer func() { outputFormat = "text" }()

	cmd := newGraphQLCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--query", "query { brand { name domain description } }"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "GitHub") {
		t.Errorf("stdout missing brand name")
	}
}

func TestGraphQLCmd_Text_Search(t *testing.T) {
	resetGraphQLFlags()
	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"brands":[{"name":"GitHub","domain":"github.com"},{"name":"GitLab","domain":"gitlab.com"}]}`)
			return json.RawMessage(data), nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	defer func() { outputFormat = "text" }()

	cmd := newGraphQLCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--query", "query { brands { name domain } }"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "GitHub") {
		t.Errorf("stdout missing search results")
	}
}

func TestGraphQLCmd_Text_ColorsFonts(t *testing.T) {
	resetGraphQLFlags()
	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"colors":[{"hex":"#000000","type":"dark","brightness":0}], "fonts":[{"name":"Inter","type":"body"}]}`)
			return json.RawMessage(data), nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	defer func() { outputFormat = "text" }()

	cmd := newGraphQLCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--query", "query { colors { hex type } fonts { name type } }"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "#000000") {
		t.Errorf("stdout missing colors")
	}
	if !strings.Contains(stdout.String(), "Inter") {
		t.Errorf("stdout missing fonts")
	}
}

func TestGraphQLCmd_Text_LogosLinks(t *testing.T) {
	resetGraphQLFlags()
	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			data := []byte(`{"logos":[{"type":"logo","theme":"light","url":"https://cdn.example.com/logo.svg","format":"svg"}],"links":[{"name":"Website","url":"https://example.com"}]}`)
			return json.RawMessage(data), nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	defer func() { outputFormat = "text" }()

	cmd := newGraphQLCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--query", "query { logos { type theme url format } links { name url } }"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "Logos:") {
		t.Errorf("stdout missing logos")
	}
	if !strings.Contains(stdout.String(), "Links:") {
		t.Errorf("stdout missing links")
	}
}

func TestGraphQLCmd_Stdin(t *testing.T) {
	resetGraphQLFlags()
	mock := &MockAPIClient{
		GraphQLFunc: func(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
			if !strings.Contains(query, "query FromStdin") {
				return nil, errors.New("unexpected query")
			}
			data := []byte(`{"ok":true}`)
			return json.RawMessage(data), nil
		},
	}

	var stdout bytes.Buffer
	outputFormat = "text"
	defer func() { outputFormat = "text" }()

	cmd := newGraphQLCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetIn(strings.NewReader("query FromStdin { me { id } }"))
	cmd.SetArgs([]string{"--stdin"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "\"ok\": true") {
		t.Errorf("stdout missing ok field: %s", stdout.String())
	}
}

func TestGraphQLCmd_StdinRaw(t *testing.T) {
	resetGraphQLFlags()
	mock := &MockAPIClient{
		GraphQLRawFunc: func(ctx context.Context, body io.Reader) (json.RawMessage, error) {
			data, _ := io.ReadAll(body)
			if !strings.Contains(string(data), "query Raw") {
				return nil, errors.New("unexpected raw body")
			}
			return json.RawMessage([]byte(`{"ok":true}`)), nil
		},
	}

	var stdout bytes.Buffer
	cmd := newGraphQLCmdWithClient(mock)
	cmd.SetOut(&stdout)
	cmd.SetIn(strings.NewReader(`{"query":"query Raw { me { id } }"}`))
	cmd.SetArgs([]string{"--stdin-raw"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "\"ok\": true") {
		t.Errorf("stdout missing ok field: %s", stdout.String())
	}
}

func resetGraphQLFlags() {
	graphqlQuery = ""
	graphqlQueryFile = ""
	graphqlVariables = ""
	graphqlStdin = false
	graphqlStdinRaw = false
}
