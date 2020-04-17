package store

import (
	"github.com/mjpitz/paxos/api"
	"github.com/mjpitz/paxos/internal/store/encoding"
	"github.com/mjpitz/paxos/internal/store/wal"
)

func NewProposalStore(log wal.Log) ProposalStore {
	return &proposalStore{
		log: log,
		enc: encoding.Proto,
	}
}

type ProposalStore interface {
	Last() (*api.Proposal, error)
	Append(proposal *api.Proposal) error
	Since(id uint64) ([]*api.Proposal, error)
}

type proposalStore struct {
	log wal.Log
	enc *encoding.Encoding
}

func (s *proposalStore) Last() (*api.Proposal, error) {
	data, err := s.log.Last()
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	proposal := &api.Proposal{}
	if err := s.enc.Unmarshal(data.Data, proposal); err != nil {
		return nil, err
	}
	return proposal, nil
}

func (s *proposalStore) Append(proposal *api.Proposal) error {
	data, err := s.enc.Marshal(proposal)
	if err != nil {
		return err
	}

	return s.log.Append(&wal.Entry{
		Id:   proposal.Id,
		Data: data,
	})
}

func (s *proposalStore) Since(id uint64) ([]*api.Proposal, error) {
	entries, err := s.log.Since(id)
	if err != nil {
		return nil, err
	}

	proposals := make([]*api.Proposal, len(entries))
	for i, entry := range entries {
		proposals[i] = &api.Proposal{}

		if err := s.enc.Unmarshal(entry.Data, proposals[i]); err != nil {
			return nil, err
		}
	}
	return proposals, nil
}

var _ ProposalStore = &proposalStore{}
