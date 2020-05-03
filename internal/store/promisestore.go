package store

import (
	"io"

	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/store/encoding"
	"github.com/mjpitz/paxos/internal/store/wal"
)

func NewPromiseStore(log wal.Log) PromiseStore {
	return &promiseStore{
		log: log,
		enc: encoding.Proto,
	}
}

type PromiseStore interface {
	io.Closer
	Last() (*api.Promise, error)
	Append(promise *api.Promise) error
}

type promiseStore struct {
	log wal.Log
	enc *encoding.Encoding
}

func (s *promiseStore) Close() error {
	return s.log.Close()
}

func (s *promiseStore) Last() (*api.Promise, error) {
	data, err := s.log.Last()
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	promise := &api.Promise{}
	if err := s.enc.Unmarshal(data.Data, promise); err != nil {
		return nil, err
	}
	return promise, nil
}

func (s *promiseStore) Append(promise *api.Promise) error {
	data, err := s.enc.Marshal(promise)
	if err != nil {
		return err
	}

	return s.log.Append(&wal.Entry{
		Id:   promise.Id,
		Data: data,
	})
}

var _ PromiseStore = &promiseStore{}
