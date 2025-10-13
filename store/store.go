package store

// Store is the interface for a key-value store.
type Store interface {
	Get(key string) (string, bool, error)
	Set(key, value string) error
	Del(key string) error
}
