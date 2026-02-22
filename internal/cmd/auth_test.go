package cmd

import (
	"bytes"
	"strings"
	"testing"
)

type MockSecretsStore struct {
	data map[string]string
}

func NewMockSecretsStore() *MockSecretsStore {
	return &MockSecretsStore{data: make(map[string]string)}
}

func (m *MockSecretsStore) Get(key string) (string, error) {
	v, ok := m.data[key]
	if !ok {
		return "", nil
	}
	return v, nil
}

func (m *MockSecretsStore) Set(key, value string) error {
	m.data[key] = value
	return nil
}

func (m *MockSecretsStore) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func TestAuthSetCmd_Stdin(t *testing.T) {
	store := NewMockSecretsStore()

	var stdout bytes.Buffer
	stdin := strings.NewReader("test_client_id\ntest_api_key\n")

	cmd := newAuthSetCmdWithStore(store)
	cmd.SetOut(&stdout)
	cmd.SetIn(stdin)
	cmd.SetArgs([]string{"--stdin"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if store.data["client_id"] != "test_client_id" {
		t.Errorf("client_id = %v, want test_client_id", store.data["client_id"])
	}
	if store.data["api_key"] != "test_api_key" {
		t.Errorf("api_key = %v, want test_api_key", store.data["api_key"])
	}
}

func TestAuthStatusCmd(t *testing.T) {
	store := NewMockSecretsStore()
	store.data["client_id"] = "some_id"
	store.data["api_key"] = "some_key"

	var stdout bytes.Buffer
	cmd := newAuthStatusCmdWithStore(store)
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if !containsStr(output, "configured") {
		t.Errorf("output should indicate configured status: %s", output)
	}
}

func TestAuthClearCmd(t *testing.T) {
	store := NewMockSecretsStore()
	store.data["client_id"] = "some_id"
	store.data["api_key"] = "some_key"

	var stdout bytes.Buffer
	cmd := newAuthClearCmdWithStore(store)
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, ok := store.data["client_id"]; ok {
		t.Errorf("client_id should be deleted")
	}
	if _, ok := store.data["api_key"]; ok {
		t.Errorf("api_key should be deleted")
	}
}
