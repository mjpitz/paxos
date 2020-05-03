package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/config"

	"github.com/sirupsen/logrus"

	"github.com/spf13/pflag"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

func exitIff(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	hostname, _ := os.Hostname()

	value := hostname
	clusterConfig := &config.Cluster{
		Members: []string{
			"localhost:8080",
		},
	}

	pflag.StringVar(&value, "value", value, "")
	pflag.StringSliceVar(&(clusterConfig.Members), "members", clusterConfig.Members, "")

	pflag.Parse()

	addresses := make([]resolver.Address, len(clusterConfig.Members))

	for i, member := range clusterConfig.Members {
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

		resp, err := proposer.Propose(context.Background(), &api.Value{
			Value: []byte(value),
		})

		if err != nil {
			logrus.Error(err)
			continue
		}

		logrus.Infof("PROPOSED: %s, MAJORITY: %s", value, string(resp.Value))
	}
}
