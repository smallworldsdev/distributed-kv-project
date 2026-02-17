package store

import (
	"testing"
)

func TestStore(t *testing.T) {
	s := NewStore()

	// Test Put and Get
	key := "foo"
	value := "bar"
	s.Put(key, value)

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
	s.Put(key, newValue)
	val, found = s.Get(key)
	if !found {
		t.Fatalf("Expected key %s to be found", key)
	}
	if val != newValue {
		t.Errorf("Expected value %s, got %s", newValue, val)
	}
}
