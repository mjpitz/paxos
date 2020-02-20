package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mjpitz/paxos/api"
	"sync"
	"time"
)

func newNode(store api.Store, acceptor api.AcceptorClient, generator api.IDGenerator) *Node {
	return &Node{
		proposerLock: &sync.Mutex{},
		acceptorLock: &sync.Mutex{},
		store: store,
		acceptor: acceptor,
		generator: generator,
	}
}

type Node struct {
	proposerLock *sync.Mutex
	acceptorLock *sync.Mutex

	store api.Store

	acceptor api.AcceptorClient
	generator api.IDGenerator
}

var _ api.ProposerServer = &Node{}
var _ api.AcceptorServer = &Node{}

func (node *Node) Propose(ctx context.Context, value *api.Value) (*api.EmptyMessage, error) {
	node.proposerLock.Lock()
	defer node.proposerLock.Unlock()

	val := value.GetValue()

	for attempt := 1; ; attempt++ {
		id, err := node.generator.Next()
		if err != nil {
			// fail out since we failed to retrieve an id
			return nil, err
		}

		prepare := &api.Request{
			Id: id,
			Attempt: uint32(attempt),
		}

		promise, err := node.acceptor.Prepare(ctx, prepare)
		if err != nil {
			// retry with a greater id
			continue
		}

		if accepted := promise.GetAccepted(); accepted != nil {
			val = accepted.Value
			attempt = 1
		}

		proposal := &api.Proposal{
			Id: id,
			Value: val,
		}

		if _, err := node.acceptor.Accept(ctx, proposal); err != nil {
			// retry with a greater id
			continue
		}

		if !bytes.Equal(val, value.GetValue()) {
			return nil, fmt.Errorf("lost consensus")
		}

		return &api.EmptyMessage{}, nil
	}
}

func (node *Node) Prepare(ctx context.Context, prepareAttempt *api.Request) (*api.Promise, error) {
	node.acceptorLock.Lock()
	defer node.acceptorLock.Unlock()

	lastPromise, err := node.store.LastPromise()
	if err != nil {
		return nil, err
	}

	if lastPromise != nil && prepareAttempt.Id <= lastPromise.Id {
		return nil, fmt.Errorf("proposed id is less than current Id")
	}

	var accepted *api.Proposal
	if prepareAttempt.Attempt > 1 {
		accepted, err = node.store.LastAccept()
		if err != nil {
			return nil, err
		}
	}

	promise := &api.Promise{
		Id: prepareAttempt.Id,
		Accepted: accepted,
	}

	err = node.store.RecordPromise(promise)
	return promise, err
}

func (node *Node) Accept(ctx context.Context, proposal *api.Proposal) (*api.Proposal, error) {
	node.acceptorLock.Lock()
	defer node.acceptorLock.Unlock()

	lastPromise, err := node.store.LastPromise()
	if err != nil {
		return nil, err
	}

	if lastPromise != nil && proposal.Id < lastPromise.Id {
		return nil, fmt.Errorf("proposed id is less than current Id")
	}

	if err = node.store.RecordAccept(proposal); err != nil {
		return nil, err
	}

	return proposal, nil
}

func (node *Node) Observe(req *api.Request, stream api.Acceptor_ObserveServer) error {
	lastSeen := req.Id

	for {
		proposals := node.store.AcceptsSince(lastSeen)

		for _, proposal := range proposals {
			if err := stream.Send(proposal); err != nil {
				return err
			}

			lastSeen = proposal.Id
		}

		time.Sleep(100 * time.Millisecond)
	}
}
