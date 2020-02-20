package api

type Store interface {
	LastPromise() (*Promise, error)
	LastAccept() (*Proposal, error)
	RecordPromise(promise *Promise) error
	RecordAccept(proposal *Proposal) error
	AcceptsSince(id uint64) []*Proposal
}
