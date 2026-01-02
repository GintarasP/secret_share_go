package store

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("secret not found")
	ErrBurned   = errors.New("secret already accessed")
	ErrTooLarge = errors.New("secret too large for memory limit")
	ErrRecycled = errors.New("secret recycled due to memory limits - server busy")
)

type item struct {
	id        string
	data      []byte
	createdAt time.Time
	size      int
}

// MemoryStore implements Store using a standard Map + List for FIFO eviction.
type MemoryStore struct {
	// Mutex protects map, list, and memory counters
	mu       sync.Mutex
	data     map[string]*list.Element
	order    *list.List // Front is Oldest, Back is Newest
	burned   sync.Map   // Tombstones for read secrets
	recycled sync.Map   // Tombstones for evicted secrets

	maxMemory      int64
	curMemory      int64
	createdCount   int64
	retrievedCount int64
}

// NewMemoryStore creates a new in-memory store with a max memory limit in bytes.
func NewMemoryStore(maxMemory int64) *MemoryStore {
	ms := &MemoryStore{
		data:      make(map[string]*list.Element),
		order:     list.New(),
		maxMemory: maxMemory,
	}
	// Start cleanup routine for TTL
	go ms.cleanupLoop()
	return ms
}

// Save stores the encrypted data. If memory is full, it evicts the oldest secrets.
func (s *MemoryStore) Save(id string, data []byte) error {
	size := len(data)
	if int64(size) > s.maxMemory {
		return ErrTooLarge
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Evict until we have space
	for s.curMemory+int64(size) > s.maxMemory {
		// Remove oldest (Front of list)
		elem := s.order.Front()
		if elem == nil {
			// Should be unreachable if size <= maxMemory, but safety check
			// Wait, if map is empty but size is taking all space?
			// Handled by initial check.
			// Logic: map empty, curMemory=0. 0+size <= max. Loop doesn't run.
			break
		}
		s.removeElement(elem, true) // True = Recycled
	}

	// 2. Add new item
	it := item{
		id:        id,
		data:      data,
		createdAt: time.Now(),
		size:      size,
	}

	if _, exists := s.data[id]; exists {
		return errors.New("id collision")
	}

	elem := s.order.PushBack(it)
	s.data[id] = elem
	s.curMemory += int64(size)
	s.createdCount++

	return nil
}

// Get retrieves the data and deletes it (Burn-on-Read).
func (s *MemoryStore) Get(id string) ([]byte, error) {
	// First check burned tombstones (concurrent map safe)
	if _, burned := s.burned.Load(id); burned {
		return nil, ErrBurned
	}
	if _, recycled := s.recycled.Load(id); recycled {
		return nil, ErrRecycled
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	elem, ok := s.data[id]
	if !ok {
		return nil, ErrNotFound
	}

	it := elem.Value.(item)

	// Remove from storage (Normal Burn)
	s.removeElement(elem, false)
	s.retrievedCount++

	// Add to burned (Tombstone)
	s.burned.Store(id, time.Now())

	return it.data, nil
}

// removeElement removes an element from map, list, and updates memory stats.
// Must be called with lock held.
// isRecycled = true means we record it as recycled. false means just deleted (burned later by caller).
func (s *MemoryStore) removeElement(elem *list.Element, isRecycled bool) {
	it := elem.Value.(item)
	s.order.Remove(elem)
	delete(s.data, it.id)
	s.curMemory -= int64(it.size)

	if isRecycled {
		s.recycled.Store(it.id, time.Now())
	}
}

// Stats returns current memory usage, limit, total created, and total retrieved.
func (s *MemoryStore) Stats() (used int64, limit int64, created int64, retrieved int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.curMemory, s.maxMemory, s.createdCount, s.retrievedCount
}

// cleanupLoop periodically removes expired secrets (TTL 15mins).
func (s *MemoryStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		s.prune()
	}
}

// prune removes expired items.
func (s *MemoryStore) prune() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Check from Front (Oldest)
	for {
		elem := s.order.Front()
		if elem == nil {
			break
		}

		it := elem.Value.(item)
		if now.Sub(it.createdAt) > 15*time.Minute {
			s.removeElement(elem, true) // Expired counts as recycled/gone
		} else {
			// List is sorted by time, so if this one isn't expired, next ones aren't either.
			break
		}
	}

	// Prune Burned (Tombstones)
	s.burned.Range(func(key, value interface{}) bool {
		burnedAt := value.(time.Time)
		if now.Sub(burnedAt) > 15*time.Minute {
			s.burned.Delete(key)
		}
		return true
	})

	// Prune Recycled (Tombstones)
	s.recycled.Range(func(key, value interface{}) bool {
		recycledAt := value.(time.Time)
		if now.Sub(recycledAt) > 15*time.Minute {
			s.recycled.Delete(key)
		}
		return true
	})
}
