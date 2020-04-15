package server

import (
	"context"
	"github.com/cenkalti/backoff/v4"
	"github.com/google/btree"
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/members"
	"github.com/mjpitz/paxos/internal/store"
	"github.com/sirupsen/logrus"
)

type entry struct {
	id uint64
	proposals map[string]*api.Proposal
}

var _ btree.Item = &entry{}

func (a *entry) Less(b btree.Item) bool {
	return a.id < b.(*entry).id
}

type Decision struct {
	server string
	proposal *api.Proposal
}

func NewLearner(members map[string]api.AcceptorClient, acceptLog store.Log) *Learner {
	return &Learner{
		members: members,
		acceptLog: acceptLog,
		tree: make(map[uint64]map[string]*api.Proposal),
	}
}

type Learner struct {
	members map[string]api.AcceptorClient
	acceptLog store.Log
	tree map[uint64]map[string]*api.Proposal
}

func (l *Learner) learnFrom(server string, member api.AcceptorClient, decisions chan *Decision) {
	backoffConfig := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      0,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}

	backoffConfig.Reset()

	_ = backoff.Retry(func () error {
		id := uint64(0)
		e, _ := l.acceptLog.Last()
		if e != nil {
			id = e.GetId()
		}

		stream, err := member.Observe(context.Background(), &api.Request{
			Id: id,
		})
		if err != nil {
			return err
		}

		for {
			proposal, err := stream.Recv()
			if err != nil {
				return err
			}

			decisions <- &Decision{
				server:   server,
				proposal: proposal,
			}
		}
	}, backoffConfig)
}

func (l *Learner) Learn(stop chan struct{}) {
	decisions := make(chan *Decision, len(l.members))

	for server, member := range l.members {
		go l.learnFrom(server, member, decisions)
	}

	majority := members.Majority(l.members)

	for {
		select {
		case <-stop:
			return
		case decision := <-decisions:
			proposal := decision.proposal
			server := decision.server

			if _, ok := l.tree[proposal.Id]; !ok {
				l.tree[proposal.Id] = make(map[string]*api.Proposal)
			}

			l.tree[proposal.Id][server] = proposal

			if len(l.tree[proposal.Id]) == majority {
				logrus.Infof("DECISION %d, %s", proposal.Id, string(proposal.Value))

				// record proposal

				if err := l.acceptLog.Append(proposal); err != nil {
					// log error
					logrus.Error(err)
					continue
				}

				// clean up all ids up to proposal
				// ordered log, ordered response, ordered chan
				// this means that all ids up to idx were not accepted

				keysToDelete := make([]uint64, 0)
				for key := range l.tree {
					if key <= proposal.Id {
						keysToDelete = append(keysToDelete, key)
					}
				}

				for _, key := range keysToDelete {
					delete(l.tree, key)
				}
			}
		}
	}
}
