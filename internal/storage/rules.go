package storage

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/pebble/v2"
	"github.com/mantis-dns/mantis/internal/domain"
)

// RuleStore implements domain.CustomRuleRepository.
type RuleStore struct {
	db *pebble.DB
}

// NewRuleStore creates a new custom rule store.
func NewRuleStore(db *pebble.DB) *RuleStore {
	return &RuleStore{db: db}
}

func (s *RuleStore) List(_ context.Context) ([]domain.CustomRule, error) {
	return s.scan(func(_ domain.CustomRule) bool { return true })
}

func (s *RuleStore) ListByType(_ context.Context, ruleType domain.RuleType) ([]domain.CustomRule, error) {
	return s.scan(func(r domain.CustomRule) bool { return r.Type == ruleType })
}

func (s *RuleStore) scan(match func(domain.CustomRule) bool) ([]domain.CustomRule, error) {
	iter, err := s.db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(PrefixRule),
		UpperBound: []byte(prefixEnd(PrefixRule)),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var result []domain.CustomRule
	for iter.First(); iter.Valid(); iter.Next() {
		var rule domain.CustomRule
		if err := json.Unmarshal(iter.Value(), &rule); err != nil {
			continue
		}
		if match(rule) {
			result = append(result, rule)
		}
	}
	return result, nil
}

func (s *RuleStore) Create(_ context.Context, rule *domain.CustomRule) error {
	data, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	return s.db.Set(RuleKey(rule.ID), data, pebble.Sync)
}

func (s *RuleStore) Delete(_ context.Context, id string) error {
	return s.db.Delete(RuleKey(id), pebble.Sync)
}
