package wal_test

import (
	"os"
	"testing"

	"github.com/mjpitz/paxos/internal/store/wal"

	"github.com/stretchr/testify/require"

	"go.etcd.io/bbolt"
)

func TestBoltDB(t *testing.T) {
	path := "test.boltdb"

	db, err := bbolt.Open(path, 0644, bbolt.DefaultOptions)
	require.Nil(t, err)

	defer os.Remove(path)

	s, err := wal.BoltDB(db)

	require.Nil(t, err)

	e2e(t, s)
}
