package wal

import (
	"io"

	"github.com/google/btree"
)

type Entry struct {
	Id   uint64
	Data []byte
}

func (a *Entry) Less(b btree.Item) bool {
	return a.Id < b.(*Entry).Id
}

var _ btree.Item = &Entry{}

type Log interface {
	io.Closer
	Last() (*Entry, error)
	Append(entry *Entry) error
	Since(id uint64) ([]*Entry, error)
}
