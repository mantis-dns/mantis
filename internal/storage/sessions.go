package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/pebble/v2"
	"github.com/mantis-dns/mantis/internal/domain"
)

// SessionStore implements domain.SessionRepository.
type SessionStore struct {
	db *pebble.DB
}

// NewSessionStore creates a new session store.
func NewSessionStore(db *pebble.DB) *SessionStore {
	return &SessionStore{db: db}
}

func (s *SessionStore) CreateSession(_ context.Context, session *domain.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.db.Set(SessionKey(session.Token), data, pebble.Sync)
}

func (s *SessionStore) GetSession(_ context.Context, token string) (*domain.Session, error) {
	val, closer, err := s.db.Get(SessionKey(token))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()

	var session domain.Session
	if err := json.Unmarshal(val, &session); err != nil {
		return nil, err
	}

	// Enforce session expiry.
	if time.Now().After(session.ExpiresAt) {
		s.db.Delete(SessionKey(token), pebble.Sync)
		return nil, domain.ErrNotFound
	}

	return &session, nil
}

func (s *SessionStore) DeleteSession(_ context.Context, token string) error {
	return s.db.Delete(SessionKey(token), pebble.Sync)
}

func (s *SessionStore) DeleteExpiredSessions(_ context.Context) (int64, error) {
	now := time.Now()
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(PrefixSession),
		UpperBound: []byte(prefixEnd(PrefixSession)),
	})
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	batch := s.db.NewBatch()
	var count int64
	for iter.First(); iter.Valid(); iter.Next() {
		var session domain.Session
		if err := json.Unmarshal(iter.Value(), &session); err != nil {
			continue
		}
		if session.ExpiresAt.Before(now) {
			batch.Delete(iter.Key(), nil)
			count++
		}
	}
	if count > 0 {
		return count, batch.Commit(pebble.Sync)
	}
	batch.Close()
	return 0, nil
}

func (s *SessionStore) CreateAPIKey(_ context.Context, key *domain.APIKey) error {
	data, err := json.Marshal(key)
	if err != nil {
		return err
	}
	return s.db.Set(APIKeyKey(key.KeyHash), data, pebble.Sync)
}

func (s *SessionStore) GetAPIKey(_ context.Context, keyHash string) (*domain.APIKey, error) {
	val, closer, err := s.db.Get(APIKeyKey(keyHash))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	defer closer.Close()

	var key domain.APIKey
	if err := json.Unmarshal(val, &key); err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *SessionStore) ListAPIKeys(_ context.Context) ([]domain.APIKey, error) {
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(PrefixAPIKey),
		UpperBound: []byte(prefixEnd(PrefixAPIKey)),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var result []domain.APIKey
	for iter.First(); iter.Valid(); iter.Next() {
		var key domain.APIKey
		if err := json.Unmarshal(iter.Value(), &key); err != nil {
			continue
		}
		result = append(result, key)
	}
	return result, nil
}

func (s *SessionStore) DeleteAPIKey(_ context.Context, id string) error {
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(PrefixAPIKey),
		UpperBound: []byte(prefixEnd(PrefixAPIKey)),
	})
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		var key domain.APIKey
		if err := json.Unmarshal(iter.Value(), &key); err != nil {
			continue
		}
		if key.ID == id {
			return s.db.Delete(iter.Key(), pebble.Sync)
		}
	}
	return domain.ErrNotFound
}

func (s *SessionStore) UpdateAPIKeyLastUsed(_ context.Context, id string, t time.Time) error {
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(PrefixAPIKey),
		UpperBound: []byte(prefixEnd(PrefixAPIKey)),
	})
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		var key domain.APIKey
		if err := json.Unmarshal(iter.Value(), &key); err != nil {
			continue
		}
		if key.ID == id {
			key.LastUsed = t
			data, err := json.Marshal(&key)
			if err != nil {
				return err
			}
			return s.db.Set(iter.Key(), data, pebble.Sync)
		}
	}
	return domain.ErrNotFound
}
