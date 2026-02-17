package store

import (
	"sync"
)

// Store is a thread-safe in-memory key-value store.
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewStore creates a new instance of Store.
func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

// Put adds or updates a key-value pair in the store.
func (s *Store) Put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves a value by key from the store.
// It returns the value and a boolean indicating if the key exists.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}
