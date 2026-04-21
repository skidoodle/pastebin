package store

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestBoltStore(t *testing.T) {
	dbPath := "test.db"
	defer os.Remove(dbPath)

	s, err := NewBoltStore(dbPath)
	require.NoError(t, err)
	defer s.Close()

	t.Run("Set and Get", func(t *testing.T) {
		id := "id1"
		hash := "hash1"
		content := "content1"

		err := s.Set(id, hash, content)
		assert.NoError(t, err)

		p, exists, err := s.Get(id)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, content, p.Content)
	})

	t.Run("GetIDByHash", func(t *testing.T) {
		id := "id2"
		hash := "hash2"
		content := "content2"

		s.Set(id, hash, content)
		storedID, exists, err := s.GetIDByHash(hash)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, id, storedID)
	})

	t.Run("Del", func(t *testing.T) {
		id := "id3"
		s.Set(id, "h3", "c3")
		err := s.Del(id)
		assert.NoError(t, err)

		_, exists, _ := s.Get(id)
		assert.False(t, exists)
	})

	t.Run("Cleanup", func(t *testing.T) {
		s.Set("old", "oldhash", "oldcontent")
		s.Cleanup(-time.Hour)
		_, exists, _ := s.Get("old")
		assert.False(t, exists)
	})

	t.Run("Cleanup Bad Data", func(t *testing.T) {
		s.db.Update(func(tx *bbolt.Tx) error {
			return tx.Bucket(pastesBucket).Put([]byte("bad"), []byte("invalid json"))
		})
		s.Cleanup(-time.Hour)
		// Should skip without panic
	})
}

func TestNewBoltStoreError(t *testing.T) {
	// Use a directory name as file path to trigger error
	err := os.Mkdir("testdir", 0755)
	require.NoError(t, err)
	defer os.RemoveAll("testdir")

	s, err := NewBoltStore("testdir")
	assert.Error(t, err)
	assert.Nil(t, s)
}
