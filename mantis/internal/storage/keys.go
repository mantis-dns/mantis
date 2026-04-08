package storage

import "fmt"

// Key prefixes for Pebble storage.
const (
	PrefixQueryLog  = "qlog:"
	PrefixBlocklist = "blsrc:"
	PrefixRule      = "rule:"
	PrefixLease     = "lease:"
	PrefixLeaseIP   = "lease-ip:"
	PrefixSetting   = "setting:"
	PrefixSession   = "session:"
	PrefixAPIKey    = "apikey:"
	PrefixStat      = "stat:"
)

// QueryLogKey generates a time-ordered query log key.
func QueryLogKey(unixNano int64, seq int) []byte {
	return []byte(fmt.Sprintf("%s%020d:%06d", PrefixQueryLog, unixNano, seq))
}

// BlocklistKey generates a blocklist source key.
func BlocklistKey(id string) []byte {
	return []byte(PrefixBlocklist + id)
}

// RuleKey generates a custom rule key.
func RuleKey(id string) []byte {
	return []byte(PrefixRule + id)
}

// LeaseKey generates a lease key by MAC address.
func LeaseKey(mac string) []byte {
	return []byte(PrefixLease + mac)
}

// LeaseIPKey generates a lease cross-index key by IP.
func LeaseIPKey(ip string) []byte {
	return []byte(PrefixLeaseIP + ip)
}

// SettingKey generates a settings key.
func SettingKey(key string) []byte {
	return []byte(PrefixSetting + key)
}

// SessionKey generates a session key from token hash.
func SessionKey(tokenHash string) []byte {
	return []byte(PrefixSession + tokenHash)
}

// APIKeyKey generates an API key storage key from key hash.
func APIKeyKey(keyHash string) []byte {
	return []byte(PrefixAPIKey + keyHash)
}
