package store

import (
	"testing"
)

func TestMemoryStore_RecycleBehavior(t *testing.T) {
	// Initialize store with very small limit (e.g., 20 bytes) to force eviction quickly
	// "secret-data-1" is 13 bytes.
	// "secret-data-2" is 13 bytes.
	// Total 26 bytes > 20 bytes.
	s := NewMemoryStore(20)

	id1 := "id-1"
	data1 := []byte("secret-data-1")
	if err := s.Save(id1, data1); err != nil {
		t.Fatalf("Failed to save id1: %v", err)
	}

	id2 := "id-2"
	data2 := []byte("secret-data-2")

	// This save should trigger eviction of id1 because 13+13=26 > 20
	// Store logic: evict until free space.
	if err := s.Save(id2, data2); err != nil {
		t.Fatalf("Failed to save id2: %v", err)
	}

	// Now id1 should be recycled
	_, err := s.Get(id1)
	if err != ErrRecycled {
		t.Errorf("Expected ErrRecycled for id1, got %v", err)
	}

	// id2 should be present
	val, err := s.Get(id2)
	if err != nil {
		t.Errorf("Expected success for id2, got %v", err)
	}
	if string(val) != string(data2) {
		t.Errorf("Data mismatch for id2")
	}

	// id2 should now be burned
	_, err = s.Get(id2)
	if err != ErrBurned {
		t.Errorf("Expected ErrBurned for id2 after reading, got %v", err)
	}
}
