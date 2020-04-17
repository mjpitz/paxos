package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/server"
	"github.com/mjpitz/paxos/internal/store"
	"github.com/mjpitz/paxos/internal/store/wal"

	"github.com/spf13/pflag"

	"go.etcd.io/bbolt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

func runServer(config *server.Config, promiseLog store.PromiseStore, acceptLog, decisionLog store.ProposalStore, stop chan struct{}) {
	svr, err := server.New(config, promiseLog, acceptLog, decisionLog)
	if err != nil {
		panic(err)
	}

	if err := svr.Start(stop); err != nil {
		panic(err)
	}
}

func runClient(config *server.Config) {
	addresses := make([]resolver.Address, len(config.Members))

	for i, member := range config.Members {
		addresses[i] = resolver.Address{
			Addr: member,
		}
	}

	r, cleanup := manual.GenerateAndRegisterManualResolver()
	defer cleanup()

	r.InitialState(resolver.State{
		Addresses: addresses,
	})

	target := fmt.Sprintf("%s:///unused", r.Scheme())
	cc, err := grpc.Dial(target,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	if err != nil {
		panic(err)
	}

	proposer := api.NewProposerClient(cc)

	for {
		time.Sleep(5 * time.Second)

		_, _ = proposer.Propose(context.Background(), &api.Value{
			Value: []byte(fmt.Sprintf("%d", config.ServerID)),
		})
	}
}

func createLog(path string) (wal.Log, error) {
	parts := strings.Split(path, "://")

	switch parts[0] {
	case "memory":
		return wal.Memory(), nil
	case "boltdb":
		db, err := bbolt.Open(parts[1], 0644, bbolt.DefaultOptions)
		if err != nil {
			return nil, err
		}
		return wal.BoltDB(db)
	}

	return nil, fmt.Errorf("unrecognized scheme: %s", parts[0])
}

func exitIff(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	config := &server.Config{
		ServerID: 0,
		Members: []string{
			"localhost:8080",
		},
		BindNetwork: "tcp",
		BindAddress: "localhost:8080",
	}

	promiseLogPath := "boltdb://promises.log"
	acceptLogPath := "boltdb://accepts.log"
	decisionLogPath := "boltdb://decisions.log"

	pflag.Uint64Var(&(config.ServerID), "server-id", config.ServerID, "")
	pflag.StringSliceVar(&(config.Members), "members", config.Members, "")
	pflag.StringVar(&(config.BindNetwork), "bind-network", config.BindNetwork, "")
	pflag.StringVar(&(config.BindAddress), "bind-address", config.BindAddress, "")
	pflag.StringVar(&promiseLogPath, "promise-log", promiseLogPath, "")
	pflag.StringVar(&acceptLogPath, "accept-log", acceptLogPath, "")
	pflag.StringVar(&decisionLogPath, "decision-log", decisionLogPath, "")

	pflag.Parse()

	promiseWAL, err := createLog(promiseLogPath)
	exitIff(err)

	acceptWAL, err := createLog(acceptLogPath)
	exitIff(err)

	decisionWAL, err := createLog(decisionLogPath)
	exitIff(err)

	promiseLog := store.NewPromiseStore(promiseWAL)
	acceptLog := store.NewProposalStore(acceptWAL)
	decisionLog := store.NewProposalStore(decisionWAL)

	stop := make(chan struct{})
	defer close(stop)

	go runServer(config, promiseLog, acceptLog, decisionLog, stop)
	runClient(config)
}
