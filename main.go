package main

import (
	"context"
	"fmt"
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/server"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"time"
)

var charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456790"

func runServer(config *server.Config, stop chan struct{}) {
	svr, err := server.New(config)
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

func main() {
	config := &server.Config{
		ServerID: 0,
		Members: []string{
			"localhost:8080",
		},
		BindNetwork: "tcp",
		BindAddress: "localhost:8080",
	}

	pflag.Uint64Var(&(config.ServerID), "server-id", config.ServerID, "")
	pflag.StringSliceVar(&(config.Members), "members", config.Members, "")
	pflag.StringVar(&(config.BindNetwork), "bind-network", config.BindNetwork, "")
	pflag.StringVar(&(config.BindAddress), "bind-address", config.BindAddress, "")

	pflag.Parse()

	stop := make(chan struct{})
	defer close(stop)

	go runServer(config, stop)
	runClient(config)
}
