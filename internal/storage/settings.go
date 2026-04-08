package storage

import (
	"context"

	"github.com/cockroachdb/pebble/v2"
	"github.com/mantis-dns/mantis/internal/domain"
)

// SettingsStore implements domain.SettingsRepository.
type SettingsStore struct {
	db *pebble.DB
}

// NewSettingsStore creates a new settings store.
func NewSettingsStore(db *pebble.DB) *SettingsStore {
	return &SettingsStore{db: db}
}

func (s *SettingsStore) Get(_ context.Context, key string) (string, error) {
	val, closer, err := s.db.Get(SettingKey(key))
	if err != nil {
		if err == pebble.ErrNotFound {
			return "", domain.ErrNotFound
		}
		return "", err
	}
	defer closer.Close()
	return string(val), nil
}

func (s *SettingsStore) Set(_ context.Context, key string, value string) error {
	return s.db.Set(SettingKey(key), []byte(value), pebble.Sync)
}

func (s *SettingsStore) GetAll(_ context.Context) ([]domain.Setting, error) {
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(PrefixSetting),
		UpperBound: []byte(prefixEnd(PrefixSetting)),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var result []domain.Setting
	for iter.First(); iter.Valid(); iter.Next() {
		key := string(iter.Key())[len(PrefixSetting):]
		result = append(result, domain.Setting{
			Key:   key,
			Value: string(iter.Value()),
		})
	}
	return result, nil
}
