package store

import "github.com/google/btree"

type Entry interface {
	btree.Item
	GetId() uint64
}

type Log interface {
	Last() (Entry, error)
	Append(obj Entry) error
	Since(id uint64) ([]Entry, error)
}
