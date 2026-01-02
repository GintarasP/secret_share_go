package store

import (
	"testing"
)

func TestMemoryStore_BurnedBehavior(t *testing.T) {
	s := NewMemoryStore(1024 * 1024)
	id := "test-burn-id"
	data := []byte("secret-data")

	// 1. Save
	if err := s.Save(id, data); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 2. Get (First time - should succeed)
	val, err := s.Get(id)
	if err != nil {
		t.Fatalf("First Get failed: %v", err)
	}
	if string(val) != string(data) {
		t.Errorf("Data mismatch")
	}

	// 3. Get (Second time - should return ErrBurned)
	_, err = s.Get(id)
	if err != ErrBurned {
		t.Errorf("Second Get expected ErrBurned, got %v", err)
	}

	// 4. Get (Random ID - should return ErrNotFound)
	_, err = s.Get("random-non-existent-id")
	if err != ErrNotFound {
		t.Errorf("Random Get expected ErrNotFound, got %v", err)
	}
}
