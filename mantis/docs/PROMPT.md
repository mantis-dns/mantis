# Mantis -- Claude Code Implementation Prompt

## Project Overview

Mantis is a network-level ad and tracker blocker written in Go. It operates as a DNS sinkhole -- intercepting DNS queries and blocking requests to advertising/tracking/malware domains by returning null responses. Unlike Pi-hole (PHP/Shell/C/dnsmasq), Mantis is a single Go binary containing: DNS server (UDP/TCP/DoH/DoT), recursive resolver, DHCPv4 server, query logging, blocklist management, and a React admin dashboard embedded via Go's `embed` package. Deploys as a single binary or Docker container. Targets home users (Raspberry Pi) through enterprise networks (500+ devices).

## Tech Stack

| Layer       | Technology                       | Version |
|-------------|----------------------------------|---------|
| Language    | Go                               | 1.26    |
| DNS         | codeberg.org/miekg/dns           | v2.x    |
| KV Store    | github.com/cockroachdb/pebble    | v2.1+   |
| HTTP Router | github.com/go-chi/chi/v5         | v5      |
| TLS/ACME    | golang.org/x/crypto/acme         | latest  |
| Radix Tree  | github.com/hashicorp/go-immutable-radix/v2 | v2 |
| DHCP        | github.com/insomniacslk/dhcp     | latest  |
| TOML        | github.com/pelletier/go-toml/v2  | v2      |
| Logging     | github.com/rs/zerolog            | latest  |
| WebSocket   | github.com/gorilla/websocket     | v1      |
| UUID        | github.com/google/uuid           | v1      |
| Rate Limit  | golang.org/x/time/rate           | latest  |
| Testing     | github.com/stretchr/testify      | v1      |
| Frontend    | React 19 + TypeScript 5.x        | —       |
| CSS         | Tailwind CSS 4.x                 | —       |
| Build (FE)  | Vite 6.x                         | —       |
| Charts      | Recharts 2.x                     | —       |
| State       | TanStack Query 5 + Zustand 5     | —       |

## Project Structure

```
mantis/
├── cmd/mantis/main.go
├── internal/
│   ├── domain/
│   │   ├── query.go
│   │   ├── blocklist.go
│   │   ├── lease.go
│   │   ├── client.go
│   │   ├── stats.go
│   │   ├── config.go
│   │   ├── session.go
│   │   ├── repository.go
│   │   └── errors.go
│   ├── dns/
│   │   ├── server.go
│   │   ├── handler.go
│   │   └── writer.go
│   ├── doh/
│   │   ├── server.go
│   │   └── handler.go
│   ├── dot/
│   │   ├── server.go
│   │   └── handler.go
│   ├── dhcp/
│   │   ├── server.go
│   │   ├── handler.go
│   │   ├── pool.go
│   │   └── lease.go
│   ├── pipeline/
│   │   ├── handler.go
│   │   ├── chain.go
│   │   ├── cache.go
│   │   ├── gravity.go
│   │   ├── clientrule.go
│   │   └── upstream.go
│   ├── resolver/
│   │   ├── forwarder.go
│   │   ├── recursive.go
│   │   ├── circuit.go
│   │   └── cache.go
│   ├── gravity/
│   │   ├── engine.go
│   │   ├── downloader.go
│   │   ├── parser.go
│   │   ├── tree.go
│   │   └── scheduler.go
│   ├── event/
│   │   ├── bus.go
│   │   └── types.go
│   ├── api/
│   │   ├── router.go
│   │   ├── auth.go
│   │   ├── setup.go
│   │   ├── stats.go
│   │   ├── queries.go
│   │   ├── blocklists.go
│   │   ├── gravity.go
│   │   ├── rules.go
│   │   ├── dhcp.go
│   │   ├── clients.go
│   │   ├── settings.go
│   │   ├── system.go
│   │   ├── dnstest.go
│   │   ├── middleware.go
│   │   └── response.go
│   ├── storage/
│   │   ├── pebble.go
│   │   ├── keys.go
│   │   ├── querylog.go
│   │   ├── blocklist.go
│   │   ├── rules.go
│   │   ├── leases.go
│   │   ├── settings.go
│   │   └── sessions.go
│   ├── stats/
│   │   ├── aggregator.go
│   │   └── collector.go
│   ├── tls/
│   │   ├── manager.go
│   │   └── acme.go
│   ├── config/
│   │   ├── config.go
│   │   ├── loader.go
│   │   └── validate.go
│   └── web/
│       ├── embed.go
│       └── spa.go
├── web/
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── api/
│   │   │   ├── client.ts
│   │   │   ├── queries.ts
│   │   │   ├── stats.ts
│   │   │   ├── blocklists.ts
│   │   │   ├── rules.ts
│   │   │   ├── dhcp.ts
│   │   │   ├── settings.ts
│   │   │   └── websocket.ts
│   │   ├── components/
│   │   │   ├── Layout.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── DataTable.tsx
│   │   │   ├── StatCard.tsx
│   │   │   ├── Chart.tsx
│   │   │   ├── SearchInput.tsx
│   │   │   ├── Toggle.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Toast.tsx
│   │   │   └── ThemeToggle.tsx
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx
│   │   │   ├── QueryLog.tsx
│   │   │   ├── Blocklists.tsx
│   │   │   ├── Rules.tsx
│   │   │   ├── Dhcp.tsx
│   │   │   ├── Network.tsx
│   │   │   ├── Settings.tsx
│   │   │   ├── Statistics.tsx
│   │   │   └── Login.tsx
│   │   ├── store/
│   │   │   ├── authStore.ts
│   │   │   └── uiStore.ts
│   │   ├── hooks/
│   │   │   ├── useWebSocket.ts
│   │   │   └── useDebounce.ts
│   │   └── utils/
│   │       ├── format.ts
│   │       └── constants.ts
│   ├── index.html
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   └── package.json
├── configs/
│   ├── mantis.example.toml
│   └── mantis.service
├── scripts/
│   ├── install.sh
│   └── build.sh
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── tests/e2e/
│   ├── setup_test.go
│   ├── dns_test.go
│   ├── api_test.go
│   └── dhcp_test.go
├── .github/workflows/
│   ├── ci.yml
│   └── release.yml
├── go.mod
├── Makefile
├── LICENSE
└── README.md
```

## Dependencies

```bash
# Go module init
go mod init github.com/mantis-dns/mantis

# Go dependencies
go get codeberg.org/miekg/dns@latest
go get github.com/cockroachdb/pebble/v2@latest
go get github.com/go-chi/chi/v5@latest
go get golang.org/x/crypto@latest
go get golang.org/x/time@latest
go get github.com/hashicorp/go-immutable-radix/v2@latest
go get github.com/pelletier/go-toml/v2@latest
go get github.com/rs/zerolog@latest
go get github.com/gorilla/websocket@latest
go get github.com/insomniacslk/dhcp@latest
go get github.com/google/uuid@latest
go get github.com/stretchr/testify@latest
```

```bash
# Frontend (in web/ directory)
npm create vite@latest . -- --template react-ts
npm install react-router-dom @tanstack/react-query zustand recharts lucide-react
npm install -D tailwindcss @tailwindcss/vite
```

## Configuration Files

### configs/mantis.example.toml

```toml
# Mantis Configuration
# All settings can be overridden with MANTIS_ environment variables
# Example: dns.listen_address -> MANTIS_DNS_LISTEN_ADDRESS

[dns]
listen_address = "0.0.0.0:53"
upstreams = ["1.1.1.1", "8.8.8.8", "9.9.9.9"]
resolver_mode = "forward"  # "forward" or "recursive"
cache_size = 10000
blocking_mode = "null"     # "null" (0.0.0.0) or "nxdomain"

[doh]
enabled = false
listen_address = "0.0.0.0:443"

[dot]
enabled = false
listen_address = "0.0.0.0:853"

[tls]
cert_file = ""
key_file = ""
acme_enabled = false
acme_domain = ""
acme_email = ""

[dhcp]
enabled = false
interface = ""
range_start = ""
range_end = ""
lease_duration = "24h"
gateway = ""
subnet_mask = "255.255.255.0"

[api]
listen_address = "0.0.0.0:8080"
rate_limit = 60  # requests per minute per session

[storage]
data_dir = "/var/lib/mantis"

[logging]
level = "info"        # debug, info, warn, error
query_log = "all"     # all, blocked, none
retention_days = 30
privacy_mode = false

[gravity]
update_interval = "24h"
```

### Makefile

```makefile
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: dev build build-frontend build-all test lint clean

dev:
	go run -ldflags="$(LDFLAGS)" ./cmd/mantis --config configs/mantis.example.toml --data-dir ./data

build:
	go build -ldflags="$(LDFLAGS)" -o dist/mantis ./cmd/mantis

build-frontend:
	cd web && npm ci && npm run build

build-all: build-frontend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/mantis-linux-amd64 ./cmd/mantis
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/mantis-linux-arm64 ./cmd/mantis
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o dist/mantis-darwin-amd64 ./cmd/mantis
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o dist/mantis-darwin-arm64 ./cmd/mantis

test:
	go test -race ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf dist/ data/ web/dist/
```

### docker/Dockerfile

```dockerfile
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS backend
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /mantis ./cmd/mantis

FROM scratch
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /mantis /mantis
EXPOSE 53/udp 53/tcp 67/udp 68/udp 443 853 8080
VOLUME /data
ENTRYPOINT ["/mantis"]
CMD ["--config", "/data/config.toml", "--data-dir", "/data"]
```

## Implementation Order

### Step 1: Project Scaffolding

**Files:** `go.mod`, `Makefile`, `.gitignore`, `LICENSE`, `README.md`, `configs/mantis.example.toml`, `cmd/mantis/main.go`, all `internal/*/doc.go` placeholder files

Initialize Go module. Create all directories from project structure. Minimal main.go that parses `--version` flag and prints version. Makefile with dev/build/test/lint targets. .gitignore covering Go artifacts, data/, dist/, web/node_modules/, web/dist/, .env.

**Verify:** `go build ./cmd/mantis && ./mantis --version`

---

### Step 2: Core Domain Types

**Files:** All files in `internal/domain/`

Define all domain entities as Go structs with JSON tags:

```go
// internal/domain/query.go
type Transport int
const (
    TransportUDP Transport = iota
    TransportTCP
    TransportDoH
    TransportDoT
)

type Query struct {
    Domain    string    `json:"domain"`
    Type      uint16    `json:"type"`
    ClientIP  net.IP    `json:"clientIp"`
    Transport Transport `json:"transport"`
}

type Response struct {
    Answers  []dns.RR      `json:"-"`
    Blocked  bool          `json:"blocked"`
    Cached   bool          `json:"cached"`
    Upstream string        `json:"upstream,omitempty"`
    Latency  time.Duration `json:"latency"`
}
```

Define repository interfaces in `internal/domain/repository.go`:
- `QueryLogRepository`: Append, Query (with filter), DeleteBefore
- `BlocklistRepository`: List, Get, Create, Update, Delete
- `CustomRuleRepository`: List, Create, Delete, ListByType
- `LeaseRepository`: Get (by MAC), GetByIP, List, Create, Update, Delete, DeleteExpired
- `SettingsRepository`: Get, Set, GetAll
- `SessionRepository`: Create, Get, Delete, DeleteExpired

Define domain errors with codes for API mapping. Use `errors.New` pattern with sentinel errors.

**Verify:** `go build ./internal/domain/...`

---

### Step 3: Configuration System

**Files:** `internal/config/config.go`, `internal/config/loader.go`, `internal/config/validate.go`

Config struct mirrors TOML structure. Load() reads TOML file, overlays MANTIS_ env vars, overlays CLI flags. Priority: flags > env > file > defaults.

```go
type Config struct {
    DNS     DNSConfig     `toml:"dns"`
    DoH     DoHConfig     `toml:"doh"`
    DoT     DoTConfig     `toml:"dot"`
    TLS     TLSConfig     `toml:"tls"`
    DHCP    DHCPConfig    `toml:"dhcp"`
    API     APIConfig     `toml:"api"`
    Storage StorageConfig `toml:"storage"`
    Logging LoggingConfig `toml:"logging"`
    Gravity GravityConfig `toml:"gravity"`
}
```

Use pelletier/go-toml/v2 for parsing. Validate: port ranges (1-65535), IP format, non-empty required fields.

**Tests:** Load from TOML, env override, flag override, missing file uses defaults, invalid config returns error.
**Verify:** `go test ./internal/config/...`

---

### Step 4: Pebble Storage Layer

**Files:** All files in `internal/storage/`

Open Pebble database. Implement all repository interfaces.

Key encoding scheme:
- `qlog:<unix_nano_20digits>:<seq_6digits>` -> QueryLogEntry (JSON)
- `blsrc:<uuid>` -> BlocklistSource (JSON)
- `rule:<uuid>` -> CustomRule (JSON)
- `lease:<mac_normalized>` -> DhcpLease (JSON)
- `lease-ip:<ipv4_padded>` -> MAC address (cross-index)
- `setting:<dotted.key>` -> JSON value
- `session:<token_sha256>` -> Session (JSON)
- `apikey:<key_sha256>` -> APIKey (JSON)
- `stat:<bucket_type>:<timestamp>` -> StatBucket (JSON)

Query log uses batch writer:

```go
type BatchWriter struct {
    db       *pebble.DB
    batch    *pebble.Batch
    count    int
    maxBatch int           // 1000
    maxAge   time.Duration // 100ms
    mu       sync.Mutex
}
```

Flush every 1000 entries or 100ms, whichever comes first. Use time.Ticker for periodic flush.

**Tests:** QueryLog batch write + time-range query. Lease CRUD + cross-index. Settings round-trip.
**Verify:** `go test -race ./internal/storage/...`

---

### Step 5: Basic DNS Server

**Files:** `internal/dns/server.go`, `internal/dns/handler.go`, `internal/dns/writer.go`

DNS server using codeberg.org/miekg/dns. Listen on UDP and TCP. Handler converts dns.Msg to domain.Query, calls a Resolver interface, converts domain.Response back to dns.Msg. Support EDNS0 with 4096 byte buffer. Graceful shutdown: drain in-flight queries.

```go
type Server struct {
    udpServer *dns.Server
    tcpServer *dns.Server
    resolver  domain.Resolver
    logger    zerolog.Logger
}

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
    query := queryFromMsg(r, w.RemoteAddr())
    resp, err := s.resolver.Resolve(context.Background(), query)
    if err != nil {
        writeServfail(w, r)
        return
    }
    writeResponse(w, r, resp)
}
```

For now, use a stub resolver that returns SERVFAIL. Real resolvers come in Step 7.

**Verify:** `dig @localhost -p 5353 example.com` returns SERVFAIL.

---

### Step 6: DNS Response Cache

**Files:** `internal/resolver/cache.go`

LRU cache with TTL-aware eviction. Thread-safe with sync.RWMutex.

```go
type DNSCache struct {
    entries map[string]*cacheEntry
    order   *list.List // LRU order
    maxSize int
    mu      sync.RWMutex
}

type cacheEntry struct {
    response  *domain.Response
    expiresAt time.Time
    element   *list.Element
}
```

Cache key: `lowercase(domain):queryType`. Periodic sweep every 60s removes expired entries.

**Tests:** Hit/miss, TTL expiry, LRU eviction, concurrent access with -race.
**Verify:** `go test -race ./internal/resolver/...`

---

### Step 7: Upstream Forwarding with Circuit Breaker

**Files:** `internal/resolver/forwarder.go`, `internal/resolver/circuit.go`

Forwarder queries all configured upstreams in parallel, uses fastest response. Circuit breaker per upstream: opens after 5 consecutive failures, 30s cooldown, half-open allows one test request.

```go
type Forwarder struct {
    upstreams []upstream
    cache     *DNSCache
    logger    zerolog.Logger
}

type upstream struct {
    address string
    client  *dns.Client
    circuit *CircuitBreaker
}

func (f *Forwarder) Resolve(ctx context.Context, q *domain.Query) (*domain.Response, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    results := make(chan result, len(f.upstreams))
    for _, u := range f.upstreams {
        if !u.circuit.Allow() { continue }
        go func(u upstream) {
            // query upstream, send result to channel
        }(u)
    }
    // return first success, cancel rest
}
```

**Tests:** Parallel resolution, failover, circuit breaker states, all-down returns SERVFAIL.
**Verify:** `go test ./internal/resolver/...`

---

### Step 8: Gravity Engine (Blocklist Radix Tree)

**Files:** `internal/gravity/tree.go`, `internal/gravity/engine.go`

Immutable radix tree for domain lookups. Domains stored reversed ("com.example.ads") for efficient suffix matching. Atomic pointer swap on rebuild for lock-free concurrent reads.

```go
type GravityEngine struct {
    tree       atomic.Pointer[iradix.Tree[struct{}]]
    allowRules atomic.Pointer[iradix.Tree[struct{}]]
    mu         sync.Mutex // protects rebuild
}

func (g *GravityEngine) IsBlocked(domain string) bool {
    if g.allowRules.Load().Get(reverseDomain(domain)) { return false }
    return g.tree.Load().Get(reverseDomain(domain))
}

func reverseDomain(d string) string {
    // "ads.example.com" -> "com.example.ads."
}
```

Wildcard matching: `*.ads.example.com` stored as prefix "com.example.ads." -- any domain with that prefix matches.

**Tests:** Exact match, wildcard match, allowlist override, concurrent reads during rebuild.
**Verify:** `go test -race ./internal/gravity/...`

---

### Step 9: DNS Query Pipeline

**Files:** All files in `internal/pipeline/`

Chain of Responsibility pattern. Pipeline order: CacheHandler -> GravityHandler -> ClientRuleHandler -> UpstreamHandler.

```go
type QueryHandler interface {
    Handle(ctx context.Context, q *domain.Query, next QueryHandler) (*domain.Response, error)
}

type Chain struct {
    handlers []QueryHandler
}

func (c *Chain) Resolve(ctx context.Context, q *domain.Query) (*domain.Response, error) {
    return c.handlers[0].Handle(ctx, q, &chainLink{c.handlers, 1})
}
```

GravityHandler: if domain is blocked, return 0.0.0.0 (A) or :: (AAAA). CacheHandler: check cache first, store result after resolution.

**Wire main.go:** Load config -> open Pebble -> build Gravity -> build pipeline -> start DNS server -> block on SIGINT/SIGTERM.

**Verify:** `dig @localhost -p 5353 [blocked-domain]` returns 0.0.0.0. `dig @localhost -p 5353 google.com` returns real IP.

**CHECKPOINT:** At this point, Mantis resolves DNS queries with blocking and caching. `go build ./cmd/mantis` succeeds, DNS queries work.

---

### Step 10: Blocklist Downloader and Parser

**Files:** `internal/gravity/downloader.go`, `internal/gravity/parser.go`

Download blocklists via HTTP with 30s timeout, retry 3x with exponential backoff. Max 5 concurrent downloads (worker pool with semaphore channel).

Parse three formats:
- **Hosts file**: Extract domain from `0.0.0.0 domain` or `127.0.0.1 domain` lines
- **Domain only**: One domain per line, skip # comments
- **Adblock**: Extract domain from `||domain^` patterns

Skip invalid lines, count errors per source.

**Tests:** Each format with httptest server, timeout handling, retry logic.

---

### Step 11: Gravity Rebuild Orchestrator

**Files:** `internal/gravity/engine.go` (extend), `internal/gravity/scheduler.go`

Rebuild(): download all enabled BlocklistSources -> parse each -> deduplicate -> merge with custom rules -> build new radix tree -> atomic swap. Persist metadata (domain count, timestamp, status) to Pebble.

Default blocklists (shipped as constants):
- StevenBlack unified hosts (https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts)
- EasyList (https://easylist.to/easylist/easylist.txt)
- EasyPrivacy (https://easylist.to/easylist/easyprivacy.txt)
- Peter Lowe's blocklist (https://pgl.yoyo.org/adservers/serverlist.php?showintro=0&mimetype=plaintext)
- OISD blocklist (https://big.oisd.nl/domainswild)

Scheduler: time.Ticker at configured interval (default 24h), cancellable via context.

---

### Step 12: Custom Allow/Block Rules

**Files:** `internal/gravity/engine.go` (extend)

AddRule()/RemoveRule(): persist to Pebble via CustomRuleRepository, update live Gravity tree immediately. Allowlist checked before blocklist. Support exact, wildcard, regex.

---

### Step 13: Event Bus and Query Logging

**Files:** `internal/event/bus.go`, `internal/event/types.go`

Channel-based pub/sub. Non-blocking publish (drop if subscriber full). Fan-out to all subscribers.

```go
type QueryEvent struct {
    Timestamp time.Time     `json:"timestamp"`
    ClientIP  string        `json:"clientIp"`
    Domain    string        `json:"domain"`
    QueryType uint16        `json:"queryType"`
    Result    string        `json:"result"` // allowed, blocked, cached, error
    Upstream  string        `json:"upstream,omitempty"`
    LatencyUs int64         `json:"latencyUs"`
    Answer    string        `json:"answer,omitempty"`
}
```

Modify pipeline chain to publish QueryEvent after each resolution. Query log writer subscribes and batch-writes to Pebble. Privacy mode: mask last octet.

**CHECKPOINT:** Mantis blocks from real downloaded blocklists. Query log persisted. `go test -race ./...` passes.

---

### Step 14: Recursive Resolver

**Files:** `internal/resolver/recursive.go`, `internal/resolver/roothints.go`

Iterative resolution from embedded root hints. Follow NS referrals: root -> TLD -> authoritative. Cache intermediate NS records. Max recursion depth 30. Detect circular CNAME chains. Apply Gravity blocking at each step.

---

### Step 15: TLS Certificate Management

**Files:** `internal/tls/manager.go`, `internal/tls/acme.go`

Load cert from files or auto-provision via golang.org/x/crypto/acme/autocert. Shared tls.Config for DoH and DoT. TLS 1.2 minimum, 1.3 preferred.

---

### Step 16: DNS-over-HTTPS Server

**Files:** `internal/doh/server.go`, `internal/doh/handler.go`

HTTPS server on configurable port (default 443). POST: application/dns-message body (DNS wire format). GET: ?dns= base64url parameter. Uses same pipeline as standard DNS. RFC 8484 compliant.

---

### Step 17: DNS-over-TLS Server

**Files:** `internal/dot/server.go`, `internal/dot/handler.go`

TLS listener on port 853. DNS wire format inside TLS stream (2-byte length prefix per RFC 7858). TCP keepalive for persistent connections.

**CHECKPOINT:** All DNS transport modes work: UDP/TCP/DoH/DoT, forward + recursive. `go test ./...` passes.

---

### Step 18: DHCPv4 Server

**Files:** All files in `internal/dhcp/`

DHCPv4 server using insomniacslk/dhcp. IP pool from configured range. Handles DISCOVER/OFFER/REQUEST/ACK/NAK/RELEASE/INFORM. Sets Mantis as DNS server in DHCP options. Sets gateway and subnet mask from config.

---

### Step 19: DHCP Lease Persistence

**Files:** `internal/dhcp/lease.go` (extend handler and pool)

Persist leases to Pebble. Load on startup. Static leases (MAC -> IP, never expire). Lease expiry background goroutine (check every 60s).

**CHECKPOINT:** DHCP + DNS fully functional. `go test ./...` passes.

---

### Step 20: API Router and Middleware

**Files:** `internal/api/router.go`, `internal/api/middleware.go`, `internal/api/response.go`

chi v5 router with route grouping. Middleware chain: RequestID -> RealIP -> Logger -> Recoverer -> CORS -> SecurityHeaders -> [Auth + RateLimit for protected routes].

Auth middleware checks HTTP-only cookie (session) or X-Api-Key header. Rate limiter: golang.org/x/time/rate, 60 req/min per session.

Standard response helpers:

```go
func Success(w http.ResponseWriter, data interface{}) { ... }
func Error(w http.ResponseWriter, code string, message string, status int) { ... }
func Paginated(w http.ResponseWriter, data interface{}, page, perPage, total int) { ... }
```

Security headers: X-Content-Type-Options: nosniff, X-Frame-Options: DENY, CSP, Referrer-Policy.

---

### Step 21: Authentication Endpoints

**Files:** `internal/api/auth.go`, `internal/api/setup.go`

- POST /auth/setup: first-run, create admin with bcrypt(password, cost=12)
- POST /auth/login: validate bcrypt, create session (crypto/rand 256-bit), Set-Cookie (HTTP-only, Secure, SameSite=Strict), 24h sliding window
- POST /auth/logout: delete session, clear cookie
- API keys: POST /auth/apikeys (generate 64-char token, store SHA-256 hash), GET /auth/apikeys, DELETE /auth/apikeys/:id
- Login rate limit: 5/min per IP

---

### Step 22: Statistics Aggregation

**Files:** `internal/stats/collector.go`, `internal/stats/aggregator.go`

Subscribe to event bus. Maintain current-hour bucket in memory. Flush to Pebble on hour rollover (key: `stat:hour:timestamp`). Track: total queries, blocked count, cache hits, per-domain counts (top 100 min-heap), per-client counts.

---

### Step 23: Stats and Query Log API

**Files:** `internal/api/stats.go`, `internal/api/queries.go`

- GET /stats/summary: totals for 24h/7d/30d, blocked %, cache hit ratio
- GET /stats/overtime: hourly points for 24h, daily for 30d
- GET /stats/top-domains: top 10 allowed and blocked
- GET /stats/top-clients: top 10 by volume
- GET /queries?page=1&perPage=50&domain=&client=&result=&from=&to=: paginated, filtered query log

---

### Step 24: Blocklist and Rules API

**Files:** `internal/api/blocklists.go`, `internal/api/gravity.go`, `internal/api/rules.go`

Blocklist CRUD. POST /gravity/rebuild triggers async rebuild (202 Accepted). GET /gravity/status returns domain count, last rebuild, per-source stats. Rules CRUD with immediate Gravity effect.

---

### Step 25: DHCP, Clients, Settings API

**Files:** `internal/api/dhcp.go`, `internal/api/clients.go`, `internal/api/settings.go`

DHCP lease list, static lease CRUD. Client list aggregated from query log. Per-client blocking toggle. Settings get/set with validation.

---

### Step 26: WebSocket Live Query Stream

**Files:** `internal/api/queries.go` (add Stream handler)

GET /queries/stream upgrades to WebSocket (gorilla/websocket). Subscribe to event bus, forward as JSON. Ping every 30s. Buffer 1000 events, drop oldest on backpressure. Max 100 concurrent connections.

---

### Step 27: System Endpoints

**Files:** `internal/api/system.go`, `internal/api/dnstest.go`

- GET /system/health: component status (dns, storage, gravity, dhcp)
- GET /system/info: version, uptime, Go version, goroutines, memory, Pebble stats
- POST /system/restart-dns: graceful DNS restart
- GET /dns/test?domain=: resolve and return detailed result

**CHECKPOINT:** Full backend API complete. All 32 endpoints functional. `go test ./...` passes. Test all endpoints with curl.

---

### Step 28: React SPA Scaffolding

**Files:** All files in `web/` root (package.json, vite.config.ts, tailwind.config.ts, tsconfig.json, index.html, src/main.tsx, src/App.tsx)

Vite + React + TypeScript. Tailwind with dark mode (class strategy). React Router with all routes (placeholder pages). TanStack Query provider. Zustand stores.

Vite proxy config for development:
```typescript
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
});
```

**Verify:** `cd web && npm run dev` starts, `npm run build` produces dist/.

---

### Step 29: Layout and Shared Components

**Files:** All files in `web/src/components/`

Layout: fixed sidebar (240px) + scrollable content. Dark mode default. Sidebar collapses at 1024px. Mobile hamburger at 768px. DataTable with sort/pagination/skeleton. StatCard, Modal, Toast, Toggle, SearchInput, ThemeToggle.

**Colors from branding:**
- Primary: #10B981 (emerald)
- Background dark: #0F172A
- Surface dark: #1E293B
- Text dark: #F9FAFB
- Allowed: #10B981 (green)
- Blocked: #F87171 (red)
- Cached: #60A5FA (blue)

**Font:** Inter for UI, JetBrains Mono for domains/code.

---

### Step 30: Frontend Authentication

**Files:** `web/src/pages/Login.tsx`, `web/src/store/authStore.ts`, `web/src/api/client.ts`

Login page: centered card, username + password fields. Auth store: login/logout state. API client: fetch wrapper with credentials:'include'. Route guard: redirect to /login if not authenticated. Handle 401 globally.

---

### Step 31: API Client Hooks

**Files:** All files in `web/src/api/`, `web/src/hooks/`, `web/src/utils/`

TanStack Query hooks for every API endpoint. Stats auto-refetch every 10s. Optimistic updates for toggles. WebSocket hook with auto-reconnect. Format utilities for numbers (1.2K), dates (relative), domains.

**CHECKPOINT:** Frontend scaffolding complete with auth. Navigates between placeholder pages. `cd web && npm run build` succeeds.

---

### Steps 32-39: Frontend Pages

Implement all 8 pages using the shared components and API hooks:

**Step 32 - Dashboard:** 4 stat cards, 24h query chart (Recharts line), top blocked/allowed tables. Refresh every 10s.

**Step 33 - Query Log:** DataTable with timestamp/domain/client/type/result/latency. Search (debounced). Filter by result type. WebSocket live stream toggle. Inline allow/block actions. CSV export. Pagination.

**Step 34 - Blocklists:** Source table with stats. Add modal (name, URL, format). Enable/disable toggle. Delete with confirm. Gravity rebuild button with status.

**Step 35 - Custom Rules:** Block/Allow tabs. Rules table with search. Add form (domain, type, comment). Import/export text.

**Step 36 - DHCP:** Active leases table. Static leases table + add form. Pool config form. Lease time as human-readable duration.

**Step 37 - Network:** Client table (hostname, IP, MAC, query count, blocking toggle). Expandable client detail with recent queries.

**Step 38 - Settings:** Tabbed form (DNS, DHCP, Encryption, Logging, System). API key management. Password change.

**Step 39 - Statistics:** Date range picker (7d/30d/90d/custom). Queries per day bar chart. Block rate line chart. CSV export.

---

### Step 40: Embed SPA in Go Binary

**Files:** `internal/web/embed.go`, `internal/web/spa.go`

```go
//go:embed all:../../web/dist
var spaFS embed.FS

func SPAHandler() http.Handler {
    dist, _ := fs.Sub(spaFS, "web/dist")
    fileServer := http.FileServer(http.FS(dist))
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Try to serve static file, fall back to index.html for SPA routes
    })
}
```

Mount at root in router. API at /api/v1/. Static assets with cache headers.

**CHECKPOINT:** Single binary serves React SPA + API. `make build-all` produces cross-platform binaries. Full end-to-end flow works.

---

### Steps 41-42: Theme Polish and Responsive

**Step 41:** Dark mode default, light mode toggle. All components respect theme. Chart colors adapt. Focus states visible.

**Step 42:** Desktop (>1024px) full layout. Tablet (768-1024px) collapsed sidebar. Mobile (<768px) hamburger menu, stacked layouts, scrollable tables. Touch targets 44x44px.

---

### Step 43: E2E Tests

**Files:** All files in `tests/e2e/`

Start real Mantis on random ports. Test: DNS resolution (blocked/allowed/cached), API (login, add blocklist, rebuild, add rule, query log), DHCP (lease flow).

---

### Steps 44-48: Release

**Step 44:** Dockerfile (3-stage: node -> go -> scratch) + docker-compose.yml
**Step 45:** install.sh (detect arch, download, create systemd service, create mantis user)
**Step 46:** GitHub Actions CI (lint, test, build) + Release (tag -> cross-compile -> GitHub Release -> Docker push)
**Step 47:** README (features, quick start, config, Docker, build, contributing). API.md with curl examples.
**Step 48:** build.sh (cross-compile, checksums). Final main.go wiring. Graceful shutdown (SIGINT/SIGTERM).

## API Reference

| Method | Path                            | Auth     | Description                          |
|--------|---------------------------------|----------|--------------------------------------|
| POST   | /api/v1/auth/setup              | Public   | First-run admin account creation     |
| POST   | /api/v1/auth/login              | Public   | Authenticate, receive session cookie |
| POST   | /api/v1/auth/logout             | Required | Invalidate session                   |
| GET    | /api/v1/stats/summary           | Required | Dashboard summary stats              |
| GET    | /api/v1/stats/overtime          | Required | Time-series query data               |
| GET    | /api/v1/stats/top-domains       | Required | Top queried/blocked domains          |
| GET    | /api/v1/stats/top-clients       | Required | Top clients by volume                |
| GET    | /api/v1/queries                 | Required | Query log (paginated, filtered)      |
| GET    | /api/v1/queries/stream          | Required | WebSocket live query stream          |
| GET    | /api/v1/blocklists              | Required | List blocklist sources               |
| POST   | /api/v1/blocklists              | Required | Add blocklist source                 |
| PUT    | /api/v1/blocklists/:id          | Required | Update blocklist source              |
| DELETE | /api/v1/blocklists/:id          | Required | Remove blocklist source              |
| POST   | /api/v1/gravity/rebuild         | Required | Trigger Gravity rebuild              |
| GET    | /api/v1/gravity/status          | Required | Gravity stats                        |
| GET    | /api/v1/rules                   | Required | List custom rules                    |
| POST   | /api/v1/rules                   | Required | Add custom rule                      |
| DELETE | /api/v1/rules/:id               | Required | Remove custom rule                   |
| GET    | /api/v1/dhcp/leases             | Required | List DHCP leases                     |
| POST   | /api/v1/dhcp/leases/static      | Required | Add static lease                     |
| DELETE | /api/v1/dhcp/leases/static/:mac | Required | Remove static lease                  |
| GET    | /api/v1/dhcp/config             | Required | Get DHCP config                      |
| PUT    | /api/v1/dhcp/config             | Required | Update DHCP config                   |
| GET    | /api/v1/clients                 | Required | List known clients                   |
| PUT    | /api/v1/clients/:ip/blocking    | Required | Toggle per-client blocking           |
| GET    | /api/v1/settings                | Required | Get all settings                     |
| PUT    | /api/v1/settings                | Required | Update settings                      |
| GET    | /api/v1/system/info             | Required | Version, uptime, resources           |
| GET    | /api/v1/system/health           | Public   | Health check                         |
| POST   | /api/v1/system/restart-dns      | Required | Restart DNS engine                   |
| GET    | /api/v1/dns/test                | Required | Test DNS resolution                  |

**Success Response:**
```json
{"data": {...}, "meta": {"requestId": "req_abc123"}}
```

**Paginated Response:**
```json
{"data": [...], "meta": {"requestId": "req_abc123", "page": 1, "perPage": 50, "total": 1234}}
```

**Error Response:**
```json
{"error": {"code": "VALIDATION_ERROR", "message": "Invalid domain format", "requestId": "req_abc123"}}
```

## Error Handling

| Category   | HTTP | Code                   | Action                              |
|------------|------|------------------------|--------------------------------------|
| Validation | 400  | VALIDATION_ERROR       | Return field-level errors            |
| Auth       | 401  | UNAUTHORIZED           | Clear session, redirect to login     |
| Forbidden  | 403  | FORBIDDEN              | Log warning                          |
| Not Found  | 404  | NOT_FOUND              | Return resource type in message      |
| Rate Limit | 429  | RATE_LIMITED            | Include Retry-After header           |
| Internal   | 500  | INTERNAL_ERROR         | Log error, return generic message    |

DNS errors: upstream timeout -> try next upstream, all down -> SERVFAIL. Blocked -> 0.0.0.0.

## Environment Variables

All config keys available as env vars with MANTIS_ prefix. Underscores replace dots and nested keys. Examples:

| Variable                     | Config Key              | Default          |
|------------------------------|-------------------------|------------------|
| MANTIS_DNS_LISTEN_ADDRESS    | dns.listen_address      | 0.0.0.0:53       |
| MANTIS_DNS_UPSTREAMS         | dns.upstreams           | 1.1.1.1,8.8.8.8 |
| MANTIS_DNS_RESOLVER_MODE     | dns.resolver_mode       | forward          |
| MANTIS_DOH_ENABLED           | doh.enabled             | false            |
| MANTIS_DOT_ENABLED           | dot.enabled             | false            |
| MANTIS_DHCP_ENABLED          | dhcp.enabled            | false            |
| MANTIS_API_LISTEN_ADDRESS    | api.listen_address      | 0.0.0.0:8080     |
| MANTIS_DATA_DIR              | storage.data_dir        | /var/lib/mantis  |
| MANTIS_LOG_LEVEL             | logging.level           | info             |

## Quality Checks

After full implementation, verify:
- [ ] `golangci-lint run ./...` passes with 0 issues
- [ ] `go test -race ./...` passes with 0 failures
- [ ] `go build -o mantis ./cmd/mantis` produces working binary
- [ ] `cd web && npm run lint && npm run build` passes
- [ ] `make build-all` produces 4 platform binaries, each < 25MB
- [ ] `docker build -f docker/Dockerfile .` succeeds, image < 50MB
- [ ] DNS blocking works: `dig @localhost ads.example.com` returns 0.0.0.0
- [ ] Admin UI loads at http://localhost:8080
- [ ] Login flow works end-to-end
- [ ] Gravity rebuild downloads and applies real blocklists
- [ ] DoH and DoT work with valid TLS certificate
- [ ] DHCP assigns IP addresses from pool
- [ ] Query log shows in real-time via WebSocket
- [ ] Graceful shutdown on SIGINT: no dropped queries, Pebble closed cleanly
