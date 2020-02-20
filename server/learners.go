package server

import (
	"context"
	"github.com/mjpitz/paxos/api"
	"github.com/sirupsen/logrus"
	"time"
)

func learn(acceptor api.AcceptorClient, channel chan *api.Proposal, lastId uint64) {
	backoff := time.Second
	backoffCount := 0

	for {
		request := &api.Request{
			Id: lastId,
		}

		client, err := acceptor.Observe(context.Background(), request)
		if err != nil {
			// cannot establish stream
			// cannot send request
			// cannot close send
			logrus.Debug(err.Error())

			// backoff
			time.Sleep(backoff)
			backoffCount++

			if backoffCount == 3 {
				backoff = backoff * 2
				backoffCount = 0
			}

			continue
		}
		backoff = time.Second
		backoffCount = 0

		for {
			proposal, err := client.Recv()
			if err != nil {
				// cannot receive message
				logrus.Debug(err.Error())

				// connect
				break
			}

			channel <- proposal

			lastId = proposal.Id
		}
	}
}

func reduce(channel chan *api.Proposal, majority int, reduced chan *api.Proposal) {
	cache := make(map[uint64]int)

	for proposal := range channel {
		val := cache[proposal.Id]
		cache[proposal.Id] = val + 1

		if cache[proposal.Id] == majority {
			reduced <- proposal
		}
	}
}

func record(acceptor api.AcceptorClient, store api.Store, lastId uint64) {
	client, _ := acceptor.Observe(context.Background(), &api.Request{
		Id: lastId,
	})

	for {
		proposal, _ := client.Recv()

		if err := store.RecordAccept(proposal); err != nil {
			logrus.Error(err.Error())
		}
	}
}
