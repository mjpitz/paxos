package members

import (
	"context"
	"fmt"
	"github.com/mjpitz/paxos/api"
	"sync"
)

func prepare(member api.AcceptorClient, ctx context.Context, in *api.Request, promises chan *api.Promise, wg *sync.WaitGroup) {
	promise, err := member.Prepare(ctx, in)
	if err == nil {
		promises <- promise
	}
	wg.Done()
}

func Prepare(members map[string]api.AcceptorClient, ctx context.Context, in *api.Request) (*api.Promise, error) {
	wg := &sync.WaitGroup{}
	wg.Add(len(members))

	promises := make(chan *api.Promise, len(members))
	for _, member := range members {
		go prepare(member, ctx, in, promises, wg)
	}

	// wait for all requests to complete
	wg.Wait()

	if len(promises) < Majority(members) {
		return nil, fmt.Errorf("lost Majority")
	}

	var accepted *api.Proposal

	for i := len(promises) ; i > 0; i-- {
		promise := <- promises

		if promise.Accepted != nil {
			if accepted == nil || accepted.Id < promise.Accepted.Id {
				accepted = promise.Accepted
			}
		}
	}

	return &api.Promise{
		Id: in.Id,
		Accepted: accepted,
	}, nil
}
