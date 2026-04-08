package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/pebble/v2"
	"github.com/mantis-dns/mantis/internal/domain"
)

// LeaseStore implements domain.LeaseRepository with MAC and IP cross-index.
type LeaseStore struct {
	db *pebble.DB
}

// NewLeaseStore creates a new lease store.
func NewLeaseStore(db *pebble.DB) *LeaseStore {
	return &LeaseStore{db: db}
}

func (s *LeaseStore) Get(_ context.Context, mac string) (*domain.DhcpLease, error) {
	val, closer, err := s.db.Get(LeaseKey(mac))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()

	var lease domain.DhcpLease
	if err := json.Unmarshal(val, &lease); err != nil {
		return nil, err
	}
	return &lease, nil
}

func (s *LeaseStore) GetByIP(_ context.Context, ip string) (*domain.DhcpLease, error) {
	macVal, closer, err := s.db.Get(LeaseIPKey(ip))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	mac := string(macVal)
	closer.Close()

	val, closer2, err := s.db.Get(LeaseKey(mac))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	defer closer2.Close()

	var lease domain.DhcpLease
	if err := json.Unmarshal(val, &lease); err != nil {
		return nil, err
	}
	return &lease, nil
}

func (s *LeaseStore) List(_ context.Context) ([]domain.DhcpLease, error) {
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(PrefixLease),
		UpperBound: []byte(prefixEnd(PrefixLease)),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var result []domain.DhcpLease
	for iter.First(); iter.Valid(); iter.Next() {
		var lease domain.DhcpLease
		if err := json.Unmarshal(iter.Value(), &lease); err != nil {
			continue
		}
		result = append(result, lease)
	}
	return result, nil
}

func (s *LeaseStore) Create(_ context.Context, lease *domain.DhcpLease) error {
	batch := s.db.NewBatch()
	data, err := json.Marshal(lease)
	if err != nil {
		batch.Close()
		return err
	}
	if err := batch.Set(LeaseKey(lease.MAC), data, nil); err != nil {
		batch.Close()
		return err
	}
	if err := batch.Set(LeaseIPKey(lease.IP), []byte(lease.MAC), nil); err != nil {
		batch.Close()
		return err
	}
	return batch.Commit(pebble.Sync)
}

func (s *LeaseStore) Update(_ context.Context, lease *domain.DhcpLease) error {
	// Remove old IP cross-index if IP changed.
	oldVal, closer, err := s.db.Get(LeaseKey(lease.MAC))
	if err == nil {
		var old domain.DhcpLease
		if json.Unmarshal(oldVal, &old) == nil && old.IP != lease.IP {
			s.db.Delete(LeaseIPKey(old.IP), pebble.Sync)
		}
		closer.Close()
	}

	return s.Create(context.Background(), lease)
}

func (s *LeaseStore) Delete(_ context.Context, mac string) error {
	val, closer, err := s.db.Get(LeaseKey(mac))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil
		}
		return err
	}
	var lease domain.DhcpLease
	json.Unmarshal(val, &lease)
	closer.Close()

	batch := s.db.NewBatch()
	batch.Delete(LeaseKey(mac), nil)
	if lease.IP != "" {
		batch.Delete(LeaseIPKey(lease.IP), nil)
	}
	return batch.Commit(pebble.Sync)
}

func (s *LeaseStore) DeleteExpired(_ context.Context) (int64, error) {
	now := time.Now()
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(PrefixLease),
		UpperBound: []byte(prefixEnd(PrefixLease)),
	})
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	batch := s.db.NewBatch()
	var count int64
	for iter.First(); iter.Valid(); iter.Next() {
		var lease domain.DhcpLease
		if err := json.Unmarshal(iter.Value(), &lease); err != nil {
			continue
		}
		if lease.IsStatic || lease.LeaseEnd.After(now) {
			continue
		}
		batch.Delete(LeaseKey(lease.MAC), nil)
		if lease.IP != "" {
			batch.Delete(LeaseIPKey(lease.IP), nil)
		}
		count++
	}
	if count > 0 {
		return count, batch.Commit(pebble.Sync)
	}
	batch.Close()
	return 0, nil
}
