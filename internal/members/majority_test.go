package members_test

import (
	"testing"

	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/members"

	"github.com/stretchr/testify/require"
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
