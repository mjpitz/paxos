package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/members"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func NewProposer(members map[string]api.AcceptorClient, generator api.IDGenerator) *Proposer {
	return &Proposer{
		mu: &sync.Mutex{},
		members: members,
		generator: generator,
	}
}

type Proposer struct {
	mu *sync.Mutex

	members map[string]api.AcceptorClient
	generator api.IDGenerator
}

func (p *Proposer) Propose(ctx context.Context, v *api.Value) (*api.EmptyMessage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	backoffConfig := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      30 * time.Second,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}

	backoffConfig.Reset()

	val := v.GetValue()
	attempt := 0

	logrus.Infof("PROPOSE %s", string(v.Value))

	err := backoff.Retry(func() error {
		attempt++

		id, err := p.generator.Next()
		if err != nil {
			// fail out since we failed to retrieve an id
			return backoff.Permanent(err)
		}

		prepare := &api.Request{
			Id: id,
			Attempt: uint32(attempt),
		}

		promise, err := members.Prepare(p.members, ctx, prepare)
		if err != nil {
			return err
		}

		if accepted := promise.GetAccepted(); accepted != nil {
			val = accepted.Value
		}

		proposal := &api.Proposal{
			Id: id,
			Value: val,
		}

		if _, err := members.Accept(p.members, ctx, proposal); err != nil {
			return err
		}

		if !bytes.Equal(val, v.GetValue()) {
			return backoff.Permanent(fmt.Errorf("lost consensus"))
		}

		return nil
	}, backoffConfig)

	return &api.EmptyMessage{}, err
}

var _ api.ProposerServer = &Proposer{}
