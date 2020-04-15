package members_test

import (
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/members"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMajority(t *testing.T) {
	clients := map[string]api.AcceptorClient{
		"a": nil,
		"b": nil,
		"c": nil,
	}

	majority := members.Majority(clients)

	require.Equal(t, 2, majority)
}
