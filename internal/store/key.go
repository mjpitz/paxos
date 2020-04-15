package store

import "github.com/google/btree"

type Key uint64

func (a Key) GetId() uint64 {
	return uint64(a)
}

func (a Key) Less(b btree.Item) bool {
	return a.GetId() < b.(Entry).GetId()
}

var _ Entry = Key(0)
