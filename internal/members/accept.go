package members

import (
	"context"
	"fmt"
	"github.com/mjpitz/paxos/api"
	"sync"
)

func accept(member api.AcceptorClient, ctx context.Context, in *api.Proposal, proposals chan *api.Proposal, wg *sync.WaitGroup) {
	proposal, err := member.Accept(ctx, in)
	if err == nil {
		proposals <- proposal
	}
	wg.Done()
}

func Accept(members map[string]api.AcceptorClient, ctx context.Context, in *api.Proposal) (*api.Proposal, error) {
	wg := &sync.WaitGroup{}
	wg.Add(len(members))

	proposals := make(chan *api.Proposal, len(members))
	for _, member := range members {
		go accept(member, ctx, in, proposals, wg)
	}

	// wait for all requests to complete
	wg.Wait()

	if len(proposals) < Majority(members) {
		return nil, fmt.Errorf("lost Majority")
	}

	return in, nil
}
