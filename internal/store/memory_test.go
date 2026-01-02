package store

import (
	"bytes"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	s := NewMemoryStore(1024 * 1024) // 1MB for testing
	id := "test-id"
	data := []byte("encrypted-secret")

	// Test Save
	if err := s.Save(id, data); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Retrieve (First time) - Should succeed
	retrieved, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !bytes.Equal(retrieved, data) {
		t.Errorf("Got %s, want %s", retrieved, data)
	}

	// Test Retrieve (Second time) - Should fail (Burn-on-Read)
	_, err = s.Get(id)
	if err != ErrBurned {
		t.Errorf("Second Get should fail with ErrBurned, got %v", err)
	}
}

func TestMemoryStoreConcurrency(t *testing.T) {
	s := NewMemoryStore(10 * 1024 * 1024)
	id := "concurrent-test"
	data := []byte("secret")

	s.Save(id, data)

	results := make(chan bool, 100)
	concurrency := 50

	for i := 0; i < concurrency; i++ {
		go func() {
			_, err := s.Get(id)
			results <- (err == nil)
		}()
	}

	successCount := 0
	for i := 0; i < concurrency; i++ {
		if success := <-results; success {
			successCount++
		}
	}

	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful retrieval, got %d", successCount)
	}
}
