package store

// Store defines the interface for storing and retrieving secrets.
type Store interface {
	// Save stores the value and returns a unique ID.
	Save(id string, data []byte) error
	// Get retrieves the data for the given ID and removes it (Burn-on-Read).
	Get(id string) ([]byte, error)
	// Stats returns usage
	Stats() (int64, int64, int64, int64)
}
