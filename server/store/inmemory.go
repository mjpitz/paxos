package store

import (
	"github.com/google/btree"
	"github.com/mjpitz/paxos/api"
	"github.com/sirupsen/logrus"
	"sync"
)

func NewInMemoryStore() api.Store {
	return &inMemoryStore{
		mu: &sync.Mutex{},

		lastPromise: nil,
		lastAccept: nil,

		log: btree.New(2),
	}
}

type inMemoryStore struct {
	mu *sync.Mutex

	lastPromise *api.Promise
	lastAccept *api.Proposal

	log *btree.BTree
}

var _ api.Store = &inMemoryStore{}

func (r *inMemoryStore) LastPromise() (*api.Promise, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.lastPromise, nil
}

func (r *inMemoryStore) LastAccept() (*api.Proposal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.lastAccept, nil
}

func (r *inMemoryStore) RecordPromise(promise *api.Promise) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if promise.Accepted != nil {
		logrus.Infof("PROMISE %d accepted %d, '%s'",
			promise.Id, promise.Accepted.Id, string(promise.Accepted.Value))
	} else {
		logrus.Infof("PROMISE %d", promise.Id)
	}

	r.lastPromise = promise
	r.log.ReplaceOrInsert(&api.Record{
		Promise: promise,
	})

	return nil
}

func (r *inMemoryStore) RecordAccept(proposal *api.Proposal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	logrus.Infof("ACCEPT %d, '%s'", proposal.Id, string(proposal.Value))

	r.lastAccept = proposal
	r.log.ReplaceOrInsert(&api.Record{
		Accept: proposal,
	})

	return nil
}

func (r *inMemoryStore) AcceptsSince(id uint64) []*api.Proposal {
	r.mu.Lock()
	defer r.mu.Unlock()

	pivot := &api.Record{
		Accept: &api.Proposal{},
	}

	results := make([]*api.Proposal, 0)

	r.log.AscendGreaterOrEqual(pivot, func(i btree.Item) bool {
		record := i.(*api.Record)

		if accept := record.Accept; accept != nil {
			results = append(results, accept)
		}

		// max 10 per batch
		//   not a hard requirement
		//   keeps the store from being locked for too long
		return len(results) < 10
	})

	return results
}
