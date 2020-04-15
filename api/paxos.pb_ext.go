package api

import "github.com/google/btree"

var _ btree.Item = &Promise{}

func (a *Promise) Less(b btree.Item) bool {
	if promise, ok := b.(*Promise); ok {
		return a.Id < promise.Id
	} else if proposal, ok := b.(*Proposal); ok {
		return a.Id < proposal.Id
	}

	return true
}

var _ btree.Item = &Proposal{}

func (a *Proposal) Less(b btree.Item) bool {
	if promise, ok := b.(*Promise); ok {
		return a.Id < promise.Id
	} else if proposal, ok := b.(*Proposal); ok {
		return a.Id < proposal.Id
	}

	return true
}
