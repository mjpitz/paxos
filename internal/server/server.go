package server

import (
	"net"

	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/idgen"
	"github.com/mjpitz/paxos/internal/store"

	"google.golang.org/grpc"
)

type Config struct {
	ServerID    uint64
	Members     []string
	BindNetwork string
	BindAddress string
}

func New(config *Config, promiseLog store.PromiseStore, acceptLog, decisionLog store.ProposalStore) (*Server, error) {
	p, err := decisionLog.Last()
	if err != nil {
		return nil, err
	}

	start := config.ServerID
	if p != nil {
		start = p.Id
	}

	step := uint64(len(config.Members))
	offset := ((start / step) * step) + config.ServerID

	members := make(map[string]api.AcceptorClient)
	for _, member := range config.Members {
		dialOptions := []grpc.DialOption{
			grpc.WithInsecure(),
		}

		cc, err := grpc.Dial(member, dialOptions...)
		if err != nil {
			return nil, err
		}

		members[member] = api.NewAcceptorClient(cc)
	}

	idGenerator := idgen.NewSequentialIDGenerator(offset, step)

	proposer := NewProposer(members, idGenerator)
	acceptor := NewAcceptor(promiseLog, acceptLog)
	learner := NewLearner(members, decisionLog, 10)

	return &Server{
		config:   config,
		proposer: proposer,
		acceptor: acceptor,
		learner:  learner,
	}, nil
}

type Server struct {
	config   *Config
	proposer *Proposer
	acceptor *Acceptor
	learner  *Learner
}

func (s *Server) serve() {
	listener, err := net.Listen(s.config.BindNetwork, s.config.BindAddress)
	if err != nil {
		panic(err)
	}

	svr := grpc.NewServer()

	api.RegisterProposerServer(svr, s.proposer)
	api.RegisterAcceptorServer(svr, s.acceptor)

	if err := svr.Serve(listener); err != nil {
		panic(err)
	}
}

func (s *Server) Start(stopCh chan struct{}) error {
	go s.serve()
	go s.learner.Learn(stopCh)

	return nil
}
