package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/pebble/v2"
	"github.com/mantis-dns/mantis/internal/domain"
)

// BlocklistStore implements domain.BlocklistRepository.
type BlocklistStore struct {
	db *pebble.DB
}

// NewBlocklistStore creates a new blocklist store.
func NewBlocklistStore(db *pebble.DB) *BlocklistStore {
	return &BlocklistStore{db: db}
}

func (s *BlocklistStore) List(_ context.Context) ([]domain.BlocklistSource, error) {
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(PrefixBlocklist),
		UpperBound: []byte(prefixEnd(PrefixBlocklist)),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var result []domain.BlocklistSource
	for iter.First(); iter.Valid(); iter.Next() {
		var src domain.BlocklistSource
		if err := json.Unmarshal(iter.Value(), &src); err != nil {
			continue
		}
		result = append(result, src)
	}
	return result, nil
}

func (s *BlocklistStore) Get(_ context.Context, id string) (*domain.BlocklistSource, error) {
	val, closer, err := s.db.Get(BlocklistKey(id))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()

	var src domain.BlocklistSource
	if err := json.Unmarshal(val, &src); err != nil {
		return nil, fmt.Errorf("unmarshal blocklist: %w", err)
	}
	return &src, nil
}

func (s *BlocklistStore) Create(_ context.Context, source *domain.BlocklistSource) error {
	data, err := json.Marshal(source)
	if err != nil {
		return err
	}
	return s.db.Set(BlocklistKey(source.ID), data, pebble.Sync)
}

func (s *BlocklistStore) Update(_ context.Context, source *domain.BlocklistSource) error {
	_, closer, err := s.db.Get(BlocklistKey(source.ID))
	if err != nil {
		if err == pebble.ErrNotFound {
			return domain.ErrNotFound
		}
		return err
	}
	closer.Close()

	data, err := json.Marshal(source)
	if err != nil {
		return err
	}
	return s.db.Set(BlocklistKey(source.ID), data, pebble.Sync)
}

func (s *BlocklistStore) Delete(_ context.Context, id string) error {
	return s.db.Delete(BlocklistKey(id), pebble.Sync)
}
