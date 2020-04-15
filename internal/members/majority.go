package members

import "github.com/mjpitz/paxos/api"

func Majority(members map[string]api.AcceptorClient) int {
	return len(members)/2 + 1
}
