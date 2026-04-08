package domain

import (
	"context"
	"time"
)

// QueryLogRepository persists and queries DNS query log entries.
type QueryLogRepository interface {
	Append(ctx context.Context, entry *QueryLogEntry) error
	Query(ctx context.Context, filter QueryLogFilter) ([]QueryLogEntry, int, error)
	DeleteBefore(ctx context.Context, before time.Time) (int64, error)
}

// BlocklistRepository manages blocklist source metadata.
type BlocklistRepository interface {
	List(ctx context.Context) ([]BlocklistSource, error)
	Get(ctx context.Context, id string) (*BlocklistSource, error)
	Create(ctx context.Context, source *BlocklistSource) error
	Update(ctx context.Context, source *BlocklistSource) error
	Delete(ctx context.Context, id string) error
}

// CustomRuleRepository manages user-defined allow/block rules.
type CustomRuleRepository interface {
	List(ctx context.Context) ([]CustomRule, error)
	ListByType(ctx context.Context, ruleType RuleType) ([]CustomRule, error)
	Create(ctx context.Context, rule *CustomRule) error
	Delete(ctx context.Context, id string) error
}

// LeaseRepository manages DHCP lease records.
type LeaseRepository interface {
	Get(ctx context.Context, mac string) (*DhcpLease, error)
	GetByIP(ctx context.Context, ip string) (*DhcpLease, error)
	List(ctx context.Context) ([]DhcpLease, error)
	Create(ctx context.Context, lease *DhcpLease) error
	Update(ctx context.Context, lease *DhcpLease) error
	Delete(ctx context.Context, mac string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// SettingsRepository manages key-value configuration settings.
type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	GetAll(ctx context.Context) ([]Setting, error)
}

// SessionRepository manages authentication sessions and API keys.
type SessionRepository interface {
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, token string) (*Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)

	CreateAPIKey(ctx context.Context, key *APIKey) error
	GetAPIKey(ctx context.Context, keyHash string) (*APIKey, error)
	ListAPIKeys(ctx context.Context) ([]APIKey, error)
	DeleteAPIKey(ctx context.Context, id string) error
	UpdateAPIKeyLastUsed(ctx context.Context, id string, t time.Time) error
}
