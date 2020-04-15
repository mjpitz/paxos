package store_test

import (
	"github.com/mjpitz/paxos/internal/store"
	"testing"
)

func TestMemory(t *testing.T) {
	s := store.Memory()
	e2e(t, s)
}
