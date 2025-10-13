package store

import (
	"encoding/json"
	"log/slog"
	"time"

	"go.etcd.io/bbolt"
)

var pastesBucket = []byte("pastes")

// Paste represents the data stored for each paste.
type Paste struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// BoltStore is a bbolt implementation of the Store interface.
type BoltStore struct {
	db *bbolt.DB
}

// NewBoltStore creates a new BoltStore and initializes the database.
func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(pastesBucket)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &BoltStore{db: db}, nil
}

// Get retrieves a value from the store.
func (s *BoltStore) Get(key string) (string, bool, error) {
	var paste Paste
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(pastesBucket)
		val := b.Get([]byte(key))
		if val == nil {
			return nil // Not found
		}
		return json.Unmarshal(val, &paste)
	})

	if err != nil {
		return "", false, err
	}
	if paste.Content == "" {
		return "", false, nil
	}
	return paste.Content, true, nil
}

// Set adds a value to the store with a timestamp.
func (s *BoltStore) Set(key, value string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(pastesBucket)
		paste := Paste{
			Content:   value,
			CreatedAt: time.Now(),
		}
		encoded, err := json.Marshal(paste)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), encoded)
	})
}

// Del removes a value from the store.
func (s *BoltStore) Del(key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(pastesBucket)
		return b.Delete([]byte(key))
	})
}

// Cleanup iterates through all pastes and deletes those older than maxAge.
func (s *BoltStore) Cleanup(maxAge time.Duration) {
	slog.Info("running cleanup for old pastes")
	var keysToDelete [][]byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(pastesBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var paste Paste
			if err := json.Unmarshal(v, &paste); err != nil {
				slog.Error("failed to unmarshal paste during cleanup", "key", string(k), "error", err)
				continue
			}
			if time.Since(paste.CreatedAt) > maxAge {
				keysToDelete = append(keysToDelete, k)
			}
		}
		return nil
	})
	if err != nil {
		slog.Error("failed to view pastes for cleanup", "error", err)
		return
	}

	if len(keysToDelete) > 0 {
		err = s.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(pastesBucket)
			for _, k := range keysToDelete {
				if err := b.Delete(k); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			slog.Error("failed to delete old pastes", "error", err)
		} else {
			slog.Info("cleanup finished", "deleted_count", len(keysToDelete))
		}
	} else {
		slog.Info("no old pastes to delete")
	}
}

// Close closes the database connection.
func (s *BoltStore) Close() error {
	return s.db.Close()
}
