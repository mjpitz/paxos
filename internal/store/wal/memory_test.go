package wal_test

import (
	"testing"

	"github.com/mjpitz/paxos/internal/store/wal"
)

func TestMemory(t *testing.T) {
	s := wal.Memory()
	e2e(t, s)
}
