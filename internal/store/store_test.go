package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore(t *testing.T) {
	s := NewStore()

	// Test Put and Get
	key := "foo"
	value := "bar"
	if err := s.Put(key, value); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	val, found := s.Get(key)
	if !found {
		t.Fatalf("Expected key %s to be found", key)
	}
	if val != value {
		t.Errorf("Expected value %s, got %s", value, val)
	}

	// Test Get non-existent key
	_, found = s.Get("baz")
	if found {
		t.Error("Expected key baz to not be found")
	}

	// Test update
	newValue := "baz"
	if err := s.Put(key, newValue); err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	val, found = s.Get(key)
	if !found {
		t.Fatalf("Expected key %s to be found", key)
	}
	if val != newValue {
		t.Errorf("Expected value %s, got %s", newValue, val)
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "store.json")

	// 1. Create store, put data, check file existence
	s1, err := NewPersistentStore(storeFile)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	key := "persist_key"
	value := "persist_value"
	if err := s1.Put(key, value); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	if _, err := os.Stat(storeFile); os.IsNotExist(err) {
		t.Fatal("Store file was not created")
	}

	// 2. Create new store from same file, check data persistence
	s2, err := NewPersistentStore(storeFile)
	if err != nil {
		t.Fatalf("Failed to create second store: %v", err)
	}

	val, found := s2.Get(key)
	if !found {
		t.Fatal("Key was not persisted")
	}
	if val != value {
		t.Errorf("Expected value %s, got %s", value, val)
	}
}
