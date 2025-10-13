package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()

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
