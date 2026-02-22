package secrets

import (
	"errors"

	"github.com/99designs/keyring"
)

const serviceName = "brandfetch"

// ErrNotFound is returned when a key is not found in the store.
var ErrNotFound = errors.New("key not found in secrets store")

// KeyringBackend abstracts keyring operations for testing.
type KeyringBackend interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

// realKeyring wraps the actual keyring library.
type realKeyring struct {
	ring keyring.Keyring
}

func (r *realKeyring) Get(key string) (string, error) {
	item, err := r.ring.Get(key)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return string(item.Data), nil
}

func (r *realKeyring) Set(key, value string) error {
	return r.ring.Set(keyring.Item{
		Key:  key,
		Data: []byte(value),
	})
}

func (r *realKeyring) Delete(key string) error {
	return r.ring.Remove(key)
}

// Store provides secret storage operations.
type Store struct {
	ring KeyringBackend
}

// NewStore creates a new Store with real keyring backend.
func NewStore() (*Store, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: serviceName,
	})
	if err != nil {
		return nil, err
	}
	return &Store{ring: &realKeyring{ring: ring}}, nil
}

// Get retrieves a secret by key.
func (s *Store) Get(key string) (string, error) {
	return s.ring.Get(key)
}

// Set stores a secret.
func (s *Store) Set(key, value string) error {
	return s.ring.Set(key, value)
}

// Delete removes a secret.
func (s *Store) Delete(key string) error {
	return s.ring.Delete(key)
}
