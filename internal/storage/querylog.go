package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/pebble/v2"
	"github.com/mantis-dns/mantis/internal/domain"
)

// QueryLogStore implements domain.QueryLogRepository with batch writing.
type QueryLogStore struct {
	db       *pebble.DB
	batch    *pebble.Batch
	count    int
	maxBatch int
	maxAge   time.Duration
	seq      atomic.Int64
	mu       sync.Mutex
	ticker   *time.Ticker
	done     chan struct{}
}

// NewQueryLogStore creates a new query log store with batch writing.
func NewQueryLogStore(db *pebble.DB) *QueryLogStore {
	s := &QueryLogStore{
		db:       db,
		batch:    db.NewBatch(),
		maxBatch: 1000,
		maxAge:   100 * time.Millisecond,
		done:     make(chan struct{}),
	}
	s.ticker = time.NewTicker(s.maxAge)
	go s.flushLoop()
	return s
}

func (s *QueryLogStore) flushLoop() {
	for {
		select {
		case <-s.ticker.C:
			s.mu.Lock()
			s.flushLocked()
			s.mu.Unlock()
		case <-s.done:
			return
		}
	}
}

// Close flushes remaining entries and stops the background flusher.
func (s *QueryLogStore) Close() error {
	s.ticker.Stop()
	close(s.done)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushLocked()
}

func (s *QueryLogStore) flushLocked() error {
	if s.count == 0 {
		return nil
	}
	err := s.batch.Commit(pebble.Sync)
	s.batch = s.db.NewBatch()
	s.count = 0
	return err
}

// Append adds a query log entry to the batch.
func (s *QueryLogStore) Append(_ context.Context, entry *domain.QueryLogEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal query log entry: %w", err)
	}

	seq := int(s.seq.Add(1))
	key := QueryLogKey(entry.Timestamp.UnixNano(), seq)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.batch.Set(key, data, nil); err != nil {
		return err
	}
	s.count++

	if s.count >= s.maxBatch {
		return s.flushLocked()
	}
	return nil
}

// Query retrieves query log entries matching the filter.
func (s *QueryLogStore) Query(_ context.Context, filter domain.QueryLogFilter) ([]domain.QueryLogEntry, int, error) {
	// Flush pending writes first.
	s.mu.Lock()
	s.flushLocked()
	s.mu.Unlock()

	var lowerBound, upperBound []byte
	if !filter.From.IsZero() {
		lowerBound = []byte(fmt.Sprintf("%s%020d", PrefixQueryLog, filter.From.UnixNano()))
	} else {
		lowerBound = []byte(PrefixQueryLog)
	}
	if !filter.To.IsZero() {
		upperBound = []byte(fmt.Sprintf("%s%020d", PrefixQueryLog, filter.To.UnixNano()+1))
	} else {
		upperBound = []byte(prefixEnd(PrefixQueryLog))
	}

	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: lowerBound,
		UpperBound: upperBound,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("create iterator: %w", err)
	}
	defer iter.Close()

	var all []domain.QueryLogEntry
	for iter.Last(); iter.Valid(); iter.Prev() {
		var entry domain.QueryLogEntry
		if err := json.Unmarshal(iter.Value(), &entry); err != nil {
			continue
		}
		if filter.Domain != "" && entry.Domain != filter.Domain {
			continue
		}
		if filter.ClientIP != "" && entry.ClientIP != filter.ClientIP {
			continue
		}
		if filter.Result != "" && entry.Result != filter.Result {
			continue
		}
		all = append(all, entry)
	}

	total := len(all)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 50
	}

	start := (page - 1) * perPage
	if start >= total {
		return nil, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return all[start:end], total, nil
}

// DeleteBefore removes all query log entries before the given time.
func (s *QueryLogStore) DeleteBefore(_ context.Context, before time.Time) (int64, error) {
	s.mu.Lock()
	s.flushLocked()
	s.mu.Unlock()

	upper := []byte(fmt.Sprintf("%s%020d", PrefixQueryLog, before.UnixNano()))
	lower := []byte(PrefixQueryLog)

	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: lower,
		UpperBound: upper,
	})
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	var count int64
	batch := s.db.NewBatch()
	for iter.First(); iter.Valid(); iter.Next() {
		if err := batch.Delete(iter.Key(), nil); err != nil {
			batch.Close()
			return count, err
		}
		count++
	}
	if count > 0 {
		if err := batch.Commit(pebble.Sync); err != nil {
			return count, err
		}
	} else {
		batch.Close()
	}
	return count, nil
}

func prefixEnd(prefix string) string {
	if len(prefix) == 0 {
		return ""
	}
	b := []byte(prefix)
	b[len(b)-1]++
	return string(b)
}
