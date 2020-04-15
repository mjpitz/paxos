package server

import (
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/store"
	"google.golang.org/grpc"
	"net"
)

type Config struct {
	ServerID    uint64
	Members     []string
	BindNetwork string
	BindAddress string
}

func New(config *Config) (*Server, error) {
	offset := config.ServerID
	step := uint64(len(config.Members))

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

	idGenerator := api.NewSequentialIDGenerator(offset, step)
	promiseLog := store.Memory()
	acceptLog := store.Memory()

	proposer := NewProposer(members, idGenerator)
	acceptor := NewAcceptor(promiseLog, acceptLog)
	learner := NewLearner(members, acceptLog)

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

func (s *Server) Serve() {
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
	go s.Serve()
	go s.learner.Learn(stopCh)

	return nil
}
