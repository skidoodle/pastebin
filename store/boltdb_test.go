package store

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
	bbolterrors "go.etcd.io/bbolt/errors"
)

func TestBoltStore(t *testing.T) {
	dbPath := "test.db"
	defer os.Remove(dbPath)

	s, err := NewBoltStore(dbPath, nil)
	require.NoError(t, err)
	defer s.Close()

	t.Run("Open Existing Store", func(t *testing.T) {
		dbPath := "existing_store.db"
		defer os.Remove(dbPath)

		s1, err := NewBoltStore(dbPath, nil)
		require.NoError(t, err)

		err = s1.Close()
		require.NoError(t, err)

		s2, err := NewBoltStore(dbPath, nil)
		require.NoError(t, err)
		defer s2.Close()

		assert.NotNil(t, s2)
	})

	t.Run("Set and Get", func(t *testing.T) {
		id := "id1"
		hash := "hash1"
		content := "content1"

		err := s.Set(id, hash, content, nil)
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

		s.Set(id, hash, content, nil)
		storedID, exists, err := s.GetIDByHash(hash)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, id, storedID)
	})

	t.Run("Del", func(t *testing.T) {
		id := "id3"
		s.Set(id, "h3", "c3", nil)
		err := s.Del(id)
		assert.NoError(t, err)

		_, exists, _ := s.Get(id)
		assert.False(t, exists)

		_, hashExists, _ := s.GetIDByHash("h3")
		assert.False(t, hashExists)
	})

	t.Run("Cleanup", func(t *testing.T) {
		s.Set("old", "oldhash", "oldcontent", nil)
		s.Cleanup(-time.Hour)
		_, exists, _ := s.Get("old")
		assert.False(t, exists)

		_, hashExists, _ := s.GetIDByHash("oldhash")
		assert.False(t, hashExists)
	})

	t.Run("Cleanup Bad Data", func(t *testing.T) {
		s.db.Update(func(tx *bbolt.Tx) error {
			return tx.Bucket(pastesBucket).Put([]byte("bad"), []byte("invalid json"))
		})
		s.Cleanup(-time.Hour)
		// Should skip without panic
	})

	t.Run("Set Marshal Error", func(t *testing.T) {
		dbPath := "marshal_error.db"
		db, _ := bbolt.Open(dbPath, 0666, nil)
		store := &BoltStore{db: db}
		defer db.Close()
		defer os.Remove(dbPath)

		err := store.Set("id", "h", "c", map[string]interface{}{
			"foo": func() {},
		})
		assert.Error(t, err)
	})

	t.Run("Set Put Error", func(t *testing.T) {
		dbPath := "put_error.db"
		s, err := NewBoltStore(dbPath, nil)
		require.NoError(t, err)
		defer s.Close()
		defer os.Remove(dbPath)

		err = s.Set("", "hash", "content", nil)
		assert.ErrorIs(t, err, bbolterrors.ErrKeyRequired)
	})

	t.Run("Del Error", func(t *testing.T) {
		dbPath := "error_del.db"
		db, _ := bbolt.Open(dbPath, 0666, nil)
		store := &BoltStore{db: db}
		db.Close()
		err := store.Del("id")
		assert.Error(t, err)
		os.Remove(dbPath)
	})
	t.Run("Get Error", func(t *testing.T) {
		dbPath := "error_get.db"
		db, _ := bbolt.Open(dbPath, 0666, nil)
		store := &BoltStore{db: db}
		db.Close()
		_, _, err := store.Get("id")
		assert.Error(t, err)
		os.Remove(dbPath)
	})
}

func TestNewBoltStoreError(t *testing.T) {
	err := os.Mkdir("testdir", 0755)
	require.NoError(t, err)
	defer os.RemoveAll("testdir")

	s, err := NewBoltStore("testdir", nil)
	assert.Error(t, err)
	assert.Nil(t, s)

	t.Run("Bucket Creation Error", func(t *testing.T) {
		dbPath := "bucket_error.db"
		defer os.Remove(dbPath)

		originalPastesBucket := pastesBucket
		pastesBucket = []byte("")
		defer func() { pastesBucket = originalPastesBucket }()

		s, err := NewBoltStore(dbPath, nil)

		assert.ErrorIs(t, err, bbolterrors.ErrBucketNameRequired)
		assert.Nil(t, s)
	})
}
