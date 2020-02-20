package api

import "github.com/google/btree"

var _ btree.Item = &Record{}

const (
	promise = 0
	accept = 1
)

func (m *Record) ids() (uint64, uint64) {
	if m.Promise != nil {
		return m.Promise.Id, promise
	}

	return m.Accept.Id, accept
}

func (m *Record) Less(b btree.Item) bool {
	aid, atype := m.ids()
	bid, btype := b.(*Record).ids()

	if aid == bid {
		return atype < btype
	}

	return aid < bid
}
