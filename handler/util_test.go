package handler

import (
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestGenerateId(t *testing.T) {
	id, err := generateId()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(id))
}

func TestGenerateIdError(t *testing.T) {
	originalReader := rand.Reader
	defer func() { rand.Reader = originalReader }()
	rand.Reader = errorReader{}

	id, err := generateId()
	assert.Error(t, err)
	assert.Empty(t, id)
}

func TestHash(t *testing.T) {
	c := "test content"
	h1 := hash(c)
	h2 := hash(c)
	assert.Equal(t, h1, h2)
	assert.NotEmpty(t, h1)
}

func TestTimeAgo(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "0s ago", TimeAgo(now))
	assert.Equal(t, "1m ago", TimeAgo(now.Add(-61*time.Second)))
	assert.Equal(t, "1h ago", TimeAgo(now.Add(-3601*time.Second)))
	assert.Equal(t, "1d ago", TimeAgo(now.Add(-86401*time.Second)))
}
