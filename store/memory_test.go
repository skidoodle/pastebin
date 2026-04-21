package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStore(t *testing.T) {
	s := NewMemoryStore()

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
		assert.NotZero(t, p.CreatedAt)
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

	t.Run("Get Non Existent", func(t *testing.T) {
		p, exists, err := s.Get("none")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Nil(t, p)

		id, exists, err := s.GetIDByHash("none")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Empty(t, id)
	})

	t.Run("Del", func(t *testing.T) {
		id := "id3"
		s.Set(id, "h3", "c3")
		err := s.Del(id)
		assert.NoError(t, err)

		_, exists, _ := s.Get(id)
		assert.False(t, exists)
	})
}
