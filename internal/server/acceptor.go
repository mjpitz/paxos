package server

import (
	"context"
	"fmt"
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/store"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

func NewAcceptor(promiseLog, acceptLog store.Log) *Acceptor {
	return &Acceptor{
		mu: &sync.Mutex{},
		promiseLog: promiseLog,
		acceptLog: acceptLog,
		updates: make(map[api.Acceptor_ObserveServer]chan *api.Proposal),
	}
}

type Acceptor struct {
	mu *sync.Mutex

	promiseLog store.Log
	acceptLog store.Log

	updates map[api.Acceptor_ObserveServer]chan *api.Proposal
}

func (a *Acceptor) Prepare(ctx context.Context, prepareAttempt *api.Request) (*api.Promise, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	logrus.Infof("PREPARE %d", prepareAttempt.Id)

	lastPromiseEntry, err := a.promiseLog.Last()
	if err != nil {
		return nil, err
	}

	if lastPromiseEntry != nil && prepareAttempt.Id <= lastPromiseEntry.(*api.Promise).Id {
		return nil, fmt.Errorf("proposed id is less than current Id")
	}

	var accepted *api.Proposal
	if prepareAttempt.Attempt > 1 {
		lastAcceptEntry, err := a.acceptLog.Last()
		if err != nil {
			return nil, err
		}

		if lastAcceptEntry != nil {
			accepted = lastAcceptEntry.(*api.Proposal)
		}
	}

	promise := &api.Promise{
		Id: prepareAttempt.Id,
		Accepted: accepted,
	}

	logrus.Infof("PROMISE %d", promise.Id)

	return promise, a.promiseLog.Append(promise)
}

func (a *Acceptor) Accept(ctx context.Context, proposal *api.Proposal) (*api.Proposal, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	logrus.Infof("ACCEPT %d", proposal.Id)

	lastPromiseEntry, err := a.promiseLog.Last()
	if err != nil {
		return nil, err
	}

	lastPromise := lastPromiseEntry.(*api.Promise)

	if lastPromise != nil && proposal.Id < lastPromise.Id {
		return nil, fmt.Errorf("proposed id is less than current Id")
	}

	logrus.Infof("ACCEPT_REQUEST %d", proposal.Id)

	if err = a.acceptLog.Append(proposal); err != nil {
		return nil, err
	}

	for _, stream := range a.updates {
		stream <- proposal
	}

	return proposal, nil
}

func (a *Acceptor) Observe(req *api.Request, stream api.Acceptor_ObserveServer) error {
	a.mu.Lock()

	subscription := make(chan *api.Proposal, 5)
	a.updates[stream] = subscription

	defer func() {
		a.mu.Lock()
		delete(a.updates, stream)
		a.mu.Unlock()
	}()
	a.mu.Unlock()

	// read entries since req id
	entries, err := a.acceptLog.Since(req.Id)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := stream.Send(entry.(*api.Proposal)); err != nil {
			return status.Error(codes.Canceled, "stream has ended.")
		}
	}

	for {
		select {
		case proposal := <-subscription:
			if err := stream.Send(proposal); err != nil {
				return status.Error(codes.Canceled, "stream has ended.")
			}
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "stream has ended.")
		}
	}
}

var _ api.AcceptorServer = &Acceptor{}
