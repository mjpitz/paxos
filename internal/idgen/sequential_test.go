package idgen_test

import (
	"github.com/mjpitz/paxos/internal/idgen"
	"github.com/stretchr/testify/require"
	"testing"
)

func sequentialCommon(t *testing.T, start, step, numGenerations uint64) {
	gen := idgen.NewSequentialIDGenerator(start, step)

	for i := uint64(0); i < numGenerations; i++ {
		next, err := gen.Next()
		require.Nil(t, err)
		require.Equal(t, int64(start), int64(next))
		start += step
	}
}

func TestSequential_Step1(t *testing.T) {
	sequentialCommon(t, 0, 1, 10)
}

func TestSequential_Step3(t *testing.T) {
	sequentialCommon(t, 0, 3, 10)
}

func TestSequential_Step5(t *testing.T) {
	sequentialCommon(t, 0, 5, 10)
}
