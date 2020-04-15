package store_test

import (
	"testing"

	"github.com/mjpitz/paxos/internal/store"

	"github.com/stretchr/testify/require"
)

func e2e(t *testing.T, log store.Log) {
	base := uint64(1234)

	last, err := log.Last()
	require.Nil(t, err)
	require.Nil(t, last)

	err = log.Append(store.Key(base))
	require.Nil(t, err)

	last, err = log.Last()
	require.Nil(t, err)
	require.Equal(t, base, last.GetId())

	for i := 0; i < 5; i++ {
		id := base + uint64(i+1)

		err = log.Append(store.Key(id))
		require.Nil(t, err)
	}

	entries, err := log.Since(base)
	require.Nil(t, err)
	require.Len(t, entries, 5)

	for i := 0; i < 5; i++ {
		id := base + uint64(i+1)

		require.Equal(t, id, entries[i].GetId())
	}
}
