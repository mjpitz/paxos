package config

type Cluster struct {
	Members []string
}

type Server struct {
	HistorySize int
	ServerID    uint64
	BindNetwork string
	BindAddress string

	PromiseLogPath  string
	AcceptLogPath   string
	DecisionLogPath string
}
