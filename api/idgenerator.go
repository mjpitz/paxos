package api

import "sync"

type IDGenerator interface {
	Next() (id uint64, err error)
}

func NewSequentialIDGenerator(offset, step uint64) IDGenerator {
	return &sequentialIDGenerator{
		mu: &sync.Mutex{},
		offset: offset,
		step: step,
	}
}

type sequentialIDGenerator struct {
	mu *sync.Mutex
	offset uint64
	step uint64
}

func (s *sequentialIDGenerator) Next() (id uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val := s.offset
	s.offset = val + s.step
	return val, nil
}
