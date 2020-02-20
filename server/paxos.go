package server

import (
	"context"
	"github.com/mjpitz/paxos/api"
	"google.golang.org/grpc"
)

type Paxos struct {
	Acceptors []api.AcceptorClient
	DecisionLog api.Store
	IDGenerator api.IDGenerator
}

func (s *Paxos) RegisterServer(svr *grpc.Server) (api.ProposerClient, error) {
	majorityAcceptor := newMajorityAcceptor(s.Acceptors)

	node := newNode(s.DecisionLog, majorityAcceptor, s.IDGenerator)

	accept, err := s.DecisionLog.LastAccept()
	if err != nil {
		return nil, err
	}

	go record(majorityAcceptor, s.DecisionLog, accept.GetId())

	api.RegisterProposerServer(svr, node)
	api.RegisterAcceptorServer(svr, node)

	return &inProcessProposer{
		proposer: node,
	}, nil
}

//

type inProcessProposer struct {
	proposer api.ProposerServer
}

var _ api.ProposerClient = &inProcessProposer{}

func (p *inProcessProposer) Propose(ctx context.Context, in *api.Value, opts ...grpc.CallOption) (*api.EmptyMessage, error) {
	return p.proposer.Propose(ctx, in)
}
