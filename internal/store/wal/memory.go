package wal

import "github.com/google/btree"

func Memory() Log {
	return &memoryLog{
		tree: btree.New(2),
	}
}

type memoryLog struct {
	tree *btree.BTree
}

func (l *memoryLog) Last() (*Entry, error) {
	last := l.tree.Max()
	if last != nil {
		return last.(*Entry), nil
	}
	return nil, nil
}

func (l *memoryLog) Append(obj *Entry) error {
	l.tree.ReplaceOrInsert(obj)
	return nil
}

func (l *memoryLog) Since(id uint64) ([]*Entry, error) {
	results := make([]*Entry, 0)

	l.tree.AscendGreaterOrEqual(&Entry{Id: id + 1}, func(i btree.Item) bool {
		results = append(results, i.(*Entry))
		return true
	})

	return results, nil
}

var _ Log = &memoryLog{}
