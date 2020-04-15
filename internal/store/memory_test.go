package store_test

import (
	"testing"

	"github.com/mjpitz/paxos/internal/store"
)

func TestMemory(t *testing.T) {
	s := store.Memory()
	e2e(t, s)
}
