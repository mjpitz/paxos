package main

import (
	"context"
	"fmt"
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
	"math/rand"
	"time"
)

var charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456790"

func randomString() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func runServer(config *server.Config, stop chan struct{}) {
	svr, err := server.New(config)
	if err != nil {
		panic(err)
	}

	if err := svr.Start(stop); err != nil {
		panic(err)
	}
}

func main() {
	stop := make(chan struct{})

	members := []string{
		"localhost:8080",
		"localhost:8081",
		"localhost:8082",
	}

	addresses := make([]resolver.Address, len(members))

	for i, member := range members {
		logrus.Infof("starting server %s", member)
		addresses[i] = resolver.Address{
			Addr: member,
		}

		go runServer(&server.Config{
			ServerID: uint64(i),
			Members: members,
			BindNetwork: "tcp",
			BindAddress: member,
		}, stop)
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

		value := randomString()

		_, _ = proposer.Propose(context.Background(), &api.Value{
			Value: []byte(value),
		})
	}
}
