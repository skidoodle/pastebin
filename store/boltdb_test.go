package store

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

// setupTestDB creates a temporary BoltDB instance for testing and returns a cleanup function.
func setupTestDB(t *testing.T) (*BoltStore, func()) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	store, err := NewBoltStore(dbPath)
	assert.NoError(t, err)

	cleanup := func() {
		if err := store.Close(); err != nil {
			slog.Error("failed to close store", "error", err)
		}
		if err := os.RemoveAll(dir); err != nil {
			slog.Error("failed to remove temp directory", "dir", dir, "error", err)
		}
	}

	return store, cleanup
}

// TestBoltStore_SetGetDel tests the Set, Get, and Del methods of the BoltStore.
func TestBoltStore_SetGetDel(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Test Set and Get
	err := store.Set("key1", "value1")
	assert.NoError(t, err)

	val, ok, err := store.Get("key1")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// Test Get non-existent key
	_, ok, err = store.Get("non-existent")
	assert.NoError(t, err)
	assert.False(t, ok)

	// Test Delete
	err = store.Del("key1")
	assert.NoError(t, err)

	_, ok, err = store.Get("key1")
	assert.NoError(t, err)
	assert.False(t, ok)
}

// TestBoltStore_Cleanup tests the Cleanup method of the BoltStore.
func TestBoltStore_Cleanup(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Set one fresh and one stale paste
	err := store.Set("fresh", "content")
	assert.NoError(t, err)

	err = store.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(pastesBucket)
		stalePaste := Paste{
			Content:   "stale",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}
		encoded, err := json.Marshal(stalePaste)
		if err != nil {
			return err
		}
		return b.Put([]byte("stale"), encoded)
	})
	assert.NoError(t, err)

	// Cleanup pastes older than 1 hour
	store.Cleanup(1 * time.Hour)

	// Check that fresh paste still exists
	_, ok, err := store.Get("fresh")
	assert.NoError(t, err)
	assert.True(t, ok)

	// Check that stale paste was deleted
	_, ok, err = store.Get("stale")
	assert.NoError(t, err)
	assert.False(t, ok)
}
