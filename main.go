package main

import (
	"context"
	"fmt"
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/server"
	"github.com/mjpitz/paxos/server/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"net"
	"os"
	"time"
)

func propose(proposer api.ProposerClient, value string) {
	for {
		time.Sleep(time.Minute)

		proposeValue := &api.Value{
			Value: []byte(value),
		}

		logrus.Infof("PROPOSE %s", value)
		if _, err := proposer.Propose(context.Background(), proposeValue); err != nil {
			logrus.Error(err.Error())
			break
		}
	}
}

func main() {
	serverId := uint64(0)
	addresses := make([]string, 0)
	bindAddress := "0.0.0.0"
	port := 8080

	command := &cobra.Command{
		Use: "paxos",
		Short: "Starts a paxos grpc server",
		RunE: func(cmd *cobra.Command, args []string) error {


			connectParams := grpc.ConnectParams{
				Backoff: backoff.DefaultConfig,
			}

			acceptors := make([]api.AcceptorClient, len(addresses))

			for i, address := range addresses {
				cc, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithConnectParams(connectParams))
				if err != nil {
					return err
				}
				acceptors[i] = api.NewAcceptorClient(cc)
			}

			idGenerator := api.NewSequentialIDGenerator(serverId, uint64(len(addresses)))

			decisionLog := store.NewInMemoryStore()

			p := &server.Paxos{
				Acceptors: acceptors,
				DecisionLog: decisionLog,
				IDGenerator: idGenerator,
			}

			svr := grpc.NewServer()
			proposer, err := p.RegisterServer(svr)
			if err != nil {
				return err
			}

			value := fmt.Sprintf("server_%d", serverId)
			go propose(proposer, value)

			address := fmt.Sprintf("%s:%d", bindAddress, port)
			logrus.Infof("starting grpc server on %s", address)

			listener, err := net.Listen("tcp", address)
			if err != nil {
				return err
			}
			return svr.Serve(listener)
		},
	}

	flags := command.Flags()
	flags.Uint64Var(&serverId, "server-id", serverId, "The id of the server")
	flags.StringArrayVar(&addresses, "address", addresses, "The address of peers")
	flags.StringVar(&bindAddress, "bind-address", bindAddress, "The address to bind to")
	flags.IntVar(&port, "port", port, "The port to bind to")

	if err := command.Execute(); err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
}
