package store

import "time"

type Paste struct {
	Content   string                 `json:"content"`
	CreatedAt time.Time              `json:"createdAt"`
	Hash      string                 `json:"hash,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type Store interface {
	Get(id string) (*Paste, bool, error)
	GetIDByHash(hash string) (string, bool, error)
	Set(id, hash, content string, metadata map[string]interface{}) error
	Del(id string) error
}
