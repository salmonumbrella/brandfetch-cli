package secrets

import (
	"errors"
	"testing"
)

// Note: The realKeyring implementation (wrapping the actual keyring library)
// and NewStore function are not tested here because they require system keyring
// access. These would be covered by integration tests. The current test coverage
// of 20% reflects testing of the Store methods (Get, Set, Delete) which are at
// 100% coverage, while the realKeyring wrapper methods remain untested.

// MockKeyring implements a test double for keyring operations
type MockKeyring struct {
	data map[string]string
}

func NewMockKeyring() *MockKeyring {
	return &MockKeyring{data: make(map[string]string)}
}

func (m *MockKeyring) Get(key string) (string, error) {
	v, ok := m.data[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (m *MockKeyring) Set(key, value string) error {
	m.data[key] = value
	return nil
}

func (m *MockKeyring) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func TestStore_SetAndGet(t *testing.T) {
	mock := NewMockKeyring()
	store := &Store{ring: mock}

	err := store.Set("test_key", "test_value")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := store.Get("test_key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got != "test_value" {
		t.Errorf("Get() = %v, want %v", got, "test_value")
	}
}

func TestStore_GetNotFound(t *testing.T) {
	mock := NewMockKeyring()
	store := &Store{ring: mock}

	_, err := store.Get("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestStore_Delete(t *testing.T) {
	mock := NewMockKeyring()
	store := &Store{ring: mock}

	_ = store.Set("to_delete", "value")
	err := store.Delete("to_delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = store.Get("to_delete")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after Delete() error = %v, want ErrNotFound", err)
	}
}

func TestStore_SetOverwrite(t *testing.T) {
	mock := NewMockKeyring()
	store := &Store{ring: mock}

	// Set initial value
	err := store.Set("key", "value1")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Overwrite with new value
	err = store.Set("key", "value2")
	if err != nil {
		t.Fatalf("Set() overwrite error = %v", err)
	}

	// Verify new value is retrieved
	got, err := store.Get("key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got != "value2" {
		t.Errorf("Get() = %v, want %v", got, "value2")
	}
}

func TestStore_DeleteNonExistent(t *testing.T) {
	mock := NewMockKeyring()
	store := &Store{ring: mock}

	// Delete should not error on non-existent key
	err := store.Delete("nonexistent")
	if err != nil {
		t.Errorf("Delete() on non-existent key error = %v, want nil", err)
	}
}

func TestStore_GetAfterDelete(t *testing.T) {
	mock := NewMockKeyring()
	store := &Store{ring: mock}

	// Set a value
	err := store.Set("key", "value")
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Delete the key
	err = store.Delete("key")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Get should return ErrNotFound
	got, err := store.Get("key")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after Delete() error = %v, want ErrNotFound", err)
	}

	if got != "" {
		t.Errorf("Get() after Delete() = %v, want empty string", got)
	}
}

func TestStore_MultipleKeys(t *testing.T) {
	mock := NewMockKeyring()
	store := &Store{ring: mock}

	// Store multiple key-value pairs
	keys := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "value4",
	}

	for key, value := range keys {
		err := store.Set(key, value)
		if err != nil {
			t.Fatalf("Set(%q) error = %v", key, err)
		}
	}

	// Retrieve and verify all keys
	for key, want := range keys {
		got, err := store.Get(key)
		if err != nil {
			t.Fatalf("Get(%q) error = %v", key, err)
		}
		if got != want {
			t.Errorf("Get(%q) = %v, want %v", key, got, want)
		}
	}

	// Delete one key and verify others remain
	err := store.Delete("key2")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = store.Get("key2")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after Delete() error = %v, want ErrNotFound", err)
	}

	// Verify other keys still exist
	for _, key := range []string{"key1", "key3", "key4"} {
		_, err := store.Get(key)
		if err != nil {
			t.Errorf("Get(%q) after deleting key2 error = %v, want nil", key, err)
		}
	}
}

func TestStore_EmptyValue(t *testing.T) {
	mock := NewMockKeyring()
	store := &Store{ring: mock}

	// Store empty string value
	err := store.Set("empty_key", "")
	if err != nil {
		t.Fatalf("Set() with empty value error = %v", err)
	}

	// Retrieve empty value
	got, err := store.Get("empty_key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got != "" {
		t.Errorf("Get() = %v, want empty string", got)
	}
}

func TestStore_SpecialCharacters(t *testing.T) {
	mock := NewMockKeyring()
	store := &Store{ring: mock}

	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "spaces in value",
			key:   "key_with_spaces",
			value: "value with spaces",
		},
		{
			name:  "special characters in value",
			key:   "key_special",
			value: "value!@#$%^&*()_+-=[]{}|;':,.<>?/~`",
		},
		{
			name:  "unicode in value",
			key:   "key_unicode",
			value: "value with unicode: 你好 🚀 مرحبا",
		},
		{
			name:  "newlines in value",
			key:   "key_newlines",
			value: "line1\nline2\nline3",
		},
		{
			name:  "dots and dashes in key",
			key:   "key.with-special_chars",
			value: "value",
		},
		{
			name:  "json-like value",
			key:   "json_key",
			value: `{"api_key": "secret123", "token": "abc-def-ghi"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set value with special characters
			err := store.Set(tc.key, tc.value)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			// Retrieve and verify
			got, err := store.Get(tc.key)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			if got != tc.value {
				t.Errorf("Get() = %v, want %v", got, tc.value)
			}
		})
	}
}

// MockFailingKeyring simulates keyring failures for error handling tests
type MockFailingKeyring struct {
	setErr    error
	getErr    error
	deleteErr error
}

func (m *MockFailingKeyring) Get(key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	return "", ErrNotFound
}

func (m *MockFailingKeyring) Set(key, value string) error {
	return m.setErr
}

func (m *MockFailingKeyring) Delete(key string) error {
	return m.deleteErr
}

func TestStore_SetError(t *testing.T) {
	mockErr := errors.New("keyring set failed")
	mock := &MockFailingKeyring{setErr: mockErr}
	store := &Store{ring: mock}

	err := store.Set("key", "value")
	if !errors.Is(err, mockErr) {
		t.Errorf("Set() error = %v, want %v", err, mockErr)
	}
}

func TestStore_GetError(t *testing.T) {
	mockErr := errors.New("keyring get failed")
	mock := &MockFailingKeyring{getErr: mockErr}
	store := &Store{ring: mock}

	_, err := store.Get("key")
	if !errors.Is(err, mockErr) {
		t.Errorf("Get() error = %v, want %v", err, mockErr)
	}
}

func TestStore_DeleteError(t *testing.T) {
	mockErr := errors.New("keyring delete failed")
	mock := &MockFailingKeyring{deleteErr: mockErr}
	store := &Store{ring: mock}

	err := store.Delete("key")
	if !errors.Is(err, mockErr) {
		t.Errorf("Delete() error = %v, want %v", err, mockErr)
	}
}
