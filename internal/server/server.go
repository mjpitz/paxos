package server

import (
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/config"
	"github.com/mjpitz/paxos/internal/idgen"
	"github.com/mjpitz/paxos/internal/store"
	"github.com/mjpitz/paxos/internal/store/wal"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

func NewForConfig(
	clusterConfig *config.Cluster,
	serverConfig *config.Server,
) (*Server, error) {
	members := make(map[string]api.AcceptorClient)
	for _, member := range clusterConfig.Members {
		cc, err := grpc.Dial(member, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}

		members[member] = api.NewAcceptorClient(cc)
	}

	promiseWAL, err := wal.New(serverConfig.PromiseLogPath)
	if err != nil {
		return nil, err
	}

	acceptWAL, err := wal.New(serverConfig.AcceptLogPath)
	if err != nil {
		return nil, err
	}

	decisionWAL, err := wal.New(serverConfig.DecisionLogPath)
	if err != nil {
		return nil, err
	}

	promiseLog := store.NewPromiseStore(promiseWAL)
	acceptLog := store.NewProposalStore(acceptWAL)
	decisionLog := store.NewProposalStore(decisionWAL)

	p, err := decisionLog.Last()
	if err != nil {
		return nil, err
	}

	start := serverConfig.ServerID
	if p != nil {
		start = p.Id
	}

	step := uint64(len(clusterConfig.Members))
	offset := ((start / step) * step) + serverConfig.ServerID

	idGenerator := idgen.NewSequentialIDGenerator(offset, step)

	return New(members, idGenerator, promiseLog, acceptLog, decisionLog, serverConfig.HistorySize), nil
}

func New(
	members map[string]api.AcceptorClient,
	idGenerator idgen.IDGenerator,
	promiseLog store.PromiseStore,
	acceptLog, decisionLog store.ProposalStore,
	historySize int,
) *Server {
	proposer := NewProposer(members, idGenerator)
	acceptor := NewAcceptor(promiseLog, acceptLog)
	learner := NewLearner(members, decisionLog, historySize)
	stop := make(chan struct{})

	return &Server{
		promiseLog:  promiseLog,
		acceptLog:   acceptLog,
		decisionLog: decisionLog,
		proposer:    proposer,
		acceptor:    acceptor,
		learner:     learner,
		stop:        stop,
	}
}

type Server struct {
	promiseLog  store.PromiseStore
	acceptLog   store.ProposalStore
	decisionLog store.ProposalStore
	proposer    *Proposer
	acceptor    *Acceptor
	learner     *Learner
	stop        chan struct{}
}

func (s *Server) Stop() {
	close(s.stop)

	if err := s.promiseLog.Close(); err != nil {
		logrus.Errorf("failed to close promise log: %v", err)
	}

	if err := s.acceptLog.Close(); err != nil {
		logrus.Errorf("failed to close accept log: %v", err)
	}

	if err := s.decisionLog.Close(); err != nil {
		logrus.Errorf("failed to close decision log: %v", err)
	}
}

func (s *Server) Register(svr *grpc.Server) {
	api.RegisterProposerServer(svr, s.proposer)
	api.RegisterAcceptorServer(svr, s.acceptor)
}

func (s *Server) Start() {
	go s.learner.Learn(s.stop)
}
