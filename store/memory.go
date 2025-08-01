package store

import "sync"

type MemoryStore struct {
	data map[string]string
	mx   sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]string),
	}
}

func (s *MemoryStore) Get(key string) (string, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	content, exists := s.data[key]
	return content, exists
}

func (s *MemoryStore) Set(key string, value string) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.data[key] = value
}

func (s *MemoryStore) Del(key string) {
	s.mx.Lock()
	defer s.mx.Unlock()

	delete(s.data, key)
}
