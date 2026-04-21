package store

import "time"

type Paste struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store interface {
	Get(id string) (*Paste, bool, error)
	GetIDByHash(hash string) (string, bool, error)
	Set(id, hash, content string) error
	Del(id string) error
}
