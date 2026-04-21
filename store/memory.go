package store

import (
	"sync"
	"time"
)

type MemoryStore struct {
	pastes map[string]*Paste
	hashes map[string]string
	mu     sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		pastes: make(map[string]*Paste),
		hashes: make(map[string]string),
	}
}

func (s *MemoryStore) Get(id string) (*Paste, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pastes[id]
	return p, ok, nil
}

func (s *MemoryStore) GetIDByHash(hash string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.hashes[hash]
	return id, ok, nil
}

func (s *MemoryStore) Set(id, hash, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pastes[id] = &Paste{
		Content:   content,
		CreatedAt: time.Now(),
	}
	s.hashes[hash] = id
	return nil
}

func (s *MemoryStore) Del(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pastes, id)
	return nil
}
