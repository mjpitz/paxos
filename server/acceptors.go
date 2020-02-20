package server

import (
	"context"
	"fmt"
	"github.com/mjpitz/paxos/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func newMajorityAcceptor(acceptors []api.AcceptorClient) api.AcceptorClient {
	return &majorityAcceptor{
		acceptors: acceptors,
	}
}

//

type majorityStream struct {
	channel chan *api.Proposal
}

var _ api.Acceptor_ObserveClient = &majorityStream{}

func (s *majorityStream) Recv() (*api.Proposal, error) {
	return <- s.channel, nil
}

func (s *majorityStream) Header() (metadata.MD, error) {
	panic("unimplemented")
}

func (s *majorityStream) Trailer() metadata.MD {
	panic("unimplemented")
}

func (s *majorityStream) CloseSend() error {
	panic("unimplemented")
}

func (s *majorityStream) Context() context.Context {
	panic("unimplemented")
}

func (s *majorityStream) SendMsg(m interface{}) error {
	panic("unimplemented")
}

func (s *majorityStream) RecvMsg(m interface{}) error {
	panic("unimplemented")
}

//

type majorityAcceptor struct {
	acceptors []api.AcceptorClient
}

var _ api.AcceptorClient = &majorityAcceptor{}

func (a *majorityAcceptor) majority() int {
	return (len(a.acceptors) / 2) + 1
}

func (a *majorityAcceptor) Prepare(ctx context.Context, in *api.Request, opts ...grpc.CallOption) (*api.Promise, error) {
	responses := make(chan *api.Promise)

	for _, acceptor := range a.acceptors {
		go func() {
			promise, err := acceptor.Prepare(ctx, in, opts...)
			if err != nil {
				logrus.Debug(err.Error())
			}
			responses <- promise
		}()
	}

	count := 0
	var accepted *api.Proposal

	for i := 0; i < len(a.acceptors); i++ {
		response := <- responses

		if response != nil {
			count++
			if response.Accepted != nil {
				if accepted == nil || accepted.Id < response.Accepted.Id {
					accepted = response.Accepted
				}
			}
		}
	}

	if count < a.majority() {
		return nil, fmt.Errorf("lost majority")
	}

	return &api.Promise{
		Id: in.Id,
		Accepted: accepted,
	}, nil
}

func (a *majorityAcceptor) Accept(ctx context.Context, in *api.Proposal, opts ...grpc.CallOption) (*api.Proposal, error) {
	responses := make(chan *api.Proposal)

	for _, acceptor := range a.acceptors {
		go func() {
			proposal, err := acceptor.Accept(ctx, in, opts...)
			if err != nil {
				logrus.Debug(err.Error())
			}
			responses <- proposal
		}()
	}

	count := 0
	for i := 0; i < len(a.acceptors); i++ {
		response := <- responses

		if response != nil {
			count++
		}
	}

	if count < a.majority() {
		return nil, fmt.Errorf("lost majority")
	}

	return in, nil
}

func (a *majorityAcceptor) Observe(ctx context.Context, in *api.Request, opts ...grpc.CallOption) (api.Acceptor_ObserveClient, error) {
	channel := make(chan *api.Proposal)
	reduced := make(chan *api.Proposal)

	for _, acceptor := range a.acceptors {
		go learn(acceptor, channel, in.Id)
	}

	go reduce(channel, a.majority(), reduced)

	return &majorityStream{
		channel: reduced,
	}, nil
}
