package store

import (
	"encoding/json"
	"time"

	"go.etcd.io/bbolt"
)

var (
	pastesBucket = []byte("pastes")
	hashesBucket = []byte("hashes")
)

type BoltStore struct {
	db *bbolt.DB
}

func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(pastesBucket); err != nil {
			return err
		}
		_, err := tx.CreateBucketIfNotExists(hashesBucket)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &BoltStore{db: db}, nil
}

func (s *BoltStore) Get(id string) (*Paste, bool, error) {
	var paste Paste
	exists := false
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(pastesBucket)
		val := b.Get([]byte(id))
		if val == nil {
			return nil
		}
		exists = true
		return json.Unmarshal(val, &paste)
	})
	if err != nil {
		return nil, false, err
	}
	return &paste, exists, nil
}

func (s *BoltStore) GetIDByHash(hash string) (string, bool, error) {
	var id string
	exists := false
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(hashesBucket)
		val := b.Get([]byte(hash))
		if val == nil {
			return nil
		}
		exists = true
		id = string(val)
		return nil
	})
	return id, exists, err
}

func (s *BoltStore) Set(id, hash, content string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		pb := tx.Bucket(pastesBucket)
		hb := tx.Bucket(hashesBucket)

		paste := Paste{
			Content:   content,
			CreatedAt: time.Now(),
		}
		encoded, err := json.Marshal(paste)
		if err != nil {
			return err
		}

		if err := pb.Put([]byte(id), encoded); err != nil {
			return err
		}
		return hb.Put([]byte(hash), []byte(id))
	})
}

func (s *BoltStore) Del(id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(pastesBucket).Delete([]byte(id))
	})
}

func (s *BoltStore) Cleanup(maxAge time.Duration) {
	var keysToDelete [][]byte
	s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(pastesBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var p Paste
			if err := json.Unmarshal(v, &p); err == nil {
				if time.Since(p.CreatedAt) > maxAge {
					keysToDelete = append(keysToDelete, k)
				}
			}
		}
		return nil
	})

	if len(keysToDelete) > 0 {
		s.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(pastesBucket)
			for _, k := range keysToDelete {
				b.Delete(k)
			}
			return nil
		})
	}
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}
