package store

import (
	"encoding/json"
	"os"
	"sync"
)

// Store is a thread-safe in-memory key-value store.
type Store struct {
	mu       sync.RWMutex
	data     map[string]string
	filePath string
}

// NewStore creates a new instance of Store (in-memory only).
func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

// NewPersistentStore creates a new instance of Store backed by a file.
// It attempts to load existing data from the file.
func NewPersistentStore(filePath string) (*Store, error) {
	s := &Store{
		data:     make(map[string]string),
		filePath: filePath,
	}

	if _, err := os.Stat(filePath); err == nil {
		if err := s.load(); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Put adds or updates a key-value pair in the store.
func (s *Store) Put(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	
	if s.filePath != "" {
		return s.save()
	}
	return nil
}

// Get retrieves a value by key from the store.
// It returns the value and a boolean indicating if the key exists.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// GetCopy returns a copy of the current data in the store.
func (s *Store) GetCopy() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	copy := make(map[string]string)
	for k, v := range s.data {
		copy[k] = v
	}
	return copy
}

// save persists the current state to the file.
// Expects caller to hold the lock.
func (s *Store) save() error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(s.data)
}

// load reads the state from the file.
func (s *Store) load() error {
	s.mu.Lock() // Ensure exclusive access during load
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&s.data)
}
