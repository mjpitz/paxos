package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mjpitz/paxos/internal/config"
	"github.com/mjpitz/paxos/internal/server"

	"github.com/sirupsen/logrus"

	"github.com/spf13/pflag"

	"google.golang.org/grpc"
)

func exitIff(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	clusterConfig := &config.Cluster{
		Members: []string{
			"localhost:8080",
		},
	}

	serverConfig := &config.Server{
		ServerID:        0,
		BindNetwork:     "tcp",
		BindAddress:     "localhost:8080",
		PromiseLogPath:  "boltdb://logs/promises.log",
		AcceptLogPath:   "boltdb://logs/accepts.log",
		DecisionLogPath: "boltdb://logs/decisions.log",
		HistorySize:     10,
	}

	pflag.StringSliceVar(&(clusterConfig.Members), "members", clusterConfig.Members, "")
	pflag.Uint64Var(&(serverConfig.ServerID), "server-id", serverConfig.ServerID, "")
	pflag.StringVar(&(serverConfig.BindNetwork), "bind-network", serverConfig.BindNetwork, "")
	pflag.StringVar(&(serverConfig.BindAddress), "bind-address", serverConfig.BindAddress, "")
	pflag.StringVar(&(serverConfig.PromiseLogPath), "promise-log", serverConfig.PromiseLogPath, "")
	pflag.StringVar(&(serverConfig.AcceptLogPath), "accept-log", serverConfig.AcceptLogPath, "")
	pflag.StringVar(&(serverConfig.DecisionLogPath), "decision-log", serverConfig.DecisionLogPath, "")
	pflag.IntVar(&(serverConfig.HistorySize), "history-size", serverConfig.HistorySize, "")

	pflag.Parse()

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	listener, err := net.Listen(serverConfig.BindNetwork, serverConfig.BindAddress)
	exitIff(err)
	defer listener.Close()

	paxosServer, err := server.NewForConfig(clusterConfig, serverConfig)
	exitIff(err)
	defer paxosServer.Stop()

	grpcServer := grpc.NewServer()
	paxosServer.Register(grpcServer)

	go func() {
		<-signals
		logrus.Infof("received shutdown signal")

		grpcServer.Stop()
		logrus.Infof("grpc stopped")
	}()

	logrus.Infof("starting grpc server on %s://%s", serverConfig.BindNetwork, serverConfig.BindAddress)
	paxosServer.Start()

	err = grpcServer.Serve(listener)
	exitIff(err)

	logrus.Infof("shutting down")
}
