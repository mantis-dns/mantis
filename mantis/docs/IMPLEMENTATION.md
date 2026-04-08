# Mantis -- Implementation Plan

> Technical blueprint derived from SPECIFICATION.md.

## 1. Tech Stack

### 1.1 Stack Summary

| Layer       | Technology                  | Version | Rationale                                                                              |
|-------------|-----------------------------|---------|----------------------------------------------------------------------------------------|
| Language    | Go                          | 1.26    | Latest stable. Native concurrency, single binary, cross-compilation, Pi-friendly       |
| DNS Library | codeberg.org/miekg/dns      | v2.x    | De facto Go DNS library. Used by CoreDNS, Caddy. v2 is 2x faster than v1              |
| KV Store    | cockroachdb/pebble          | v2.1+   | CockroachDB's LSM engine. Write-optimized for query logging. Production-proven         |
| HTTP Router | go-chi/chi                  | v5      | Lightweight, stdlib-compatible, middleware composable. No external deps                |
| TLS/ACME    | golang.org/x/crypto/acme    | latest  | Official Go ACME client. Auto Let's Encrypt provisioning                               |
| Frontend    | React + TypeScript          | 19.x    | Large ecosystem, team familiarity. Embedded via Go embed                               |
| CSS         | Tailwind CSS                | 4.x     | Utility-first, dark mode built-in, small production bundle                             |
| Build (FE)  | Vite                        | 6.x     | Fast HMR, optimized production builds, React plugin                                    |
| Charts      | Recharts                    | 2.x     | React-native charting, lightweight, good for time-series                               |
| State (FE)  | TanStack Query + Zustand    | 5.x/5.x | Server state (Query) + client state (Zustand). Minimal boilerplate                     |
| Testing     | Go testing + testify        | —       | Stdlib test runner + assertion helpers                                                 |
| Testing (FE)| Vitest + Testing Library    | —       | Vite-native, fast, React Testing Library for components                                |
| Linting     | golangci-lint               | latest  | Meta-linter aggregating 50+ Go linters                                                 |
| Linting (FE)| ESLint + Prettier           | latest  | TypeScript linting + formatting                                                        |
| Container   | Docker (multi-stage)        | —       | Scratch-based final image, multi-arch build                                            |
| CI/CD       | GitHub Actions              | —       | Integrated with GitHub releases, matrix builds for multi-arch                          |

### 1.2 Key Technical Decisions

#### Decision: Pebble as primary storage engine

- **Context**: Mantis needs persistent storage for query logs (write-heavy, time-series), DHCP leases, blocklist metadata, settings, and custom rules (SPEC SS5).
- **Options Considered**:
  1. **SQLite**: Mature, SQL query support / CGo dependency breaks cross-compilation, single-writer bottleneck
  2. **bbolt**: Pure Go, battle-tested / B+tree optimized for reads not writes, no built-in compaction
  3. **Pebble**: Pure Go LSM, write-optimized, CockroachDB production-proven / No SQL, key-value only
- **Choice**: Pebble v2
- **Rationale**: Query logging generates thousands of writes/sec. LSM-tree's sequential write pattern matches perfectly. Pure Go enables CGo-free cross-compilation for ARM. CockroachDB uses it in production at petabyte scale.
- **Consequences**: No SQL queries -- all data access through key prefix scans and range queries. Need careful key design for time-series data.

#### Decision: miekg/dns v2 for DNS engine

- **Context**: Core DNS server functionality (SPEC SS3.1). Need full DNS protocol support including EDNS0, DNSSEC, DoH, DoT.
- **Options Considered**:
  1. **miekg/dns v2**: Complete DNS library, v2 is 2x faster / Moved to Codeberg, API not yet stable (pre-v1.0)
  2. **CoreDNS as library**: Plugin architecture / Too opinionated, heavy dependency for embedding
  3. **Custom implementation**: Full control / Enormous effort, RFC compliance risk
- **Choice**: miekg/dns v2 (codeberg.org/miekg/dns)
- **Rationale**: De facto standard for Go DNS. CoreDNS, Caddy DNS, and dozens of production systems use it. v2 offers significant performance improvements. API instability acceptable -- Mantis wraps it behind internal interfaces.
- **Consequences**: Must pin version and wrap behind interfaces to isolate API changes. Import from Codeberg, not GitHub.

#### Decision: chi v5 for HTTP API

- **Context**: Admin API with 32 endpoints (SPEC SS6.2), WebSocket support, middleware chain.
- **Options Considered**:
  1. **net/http (stdlib)**: No dependency / Verbose routing, no middleware composition in Go <1.22
  2. **chi v5**: Stdlib-compatible, zero deps, middleware chain / One more dependency
  3. **Fiber/Gin**: Performance / Not stdlib-compatible, larger dependency tree
- **Choice**: chi v5
- **Rationale**: 100% net/http compatible (handlers are standard http.HandlerFunc). Built-in middleware composition. Zero external dependencies itself. Can be swapped to stdlib later if desired.
- **Consequences**: None significant. chi adds ~500 LOC to binary.

#### Decision: Embedded React SPA via Go embed

- **Context**: Admin UI must ship inside the Go binary (SPEC SS9.2, SS7).
- **Options Considered**:
  1. **Go templates + HTMX**: Simpler / Limited interactivity for real-time query log and charts
  2. **Embedded React SPA**: Rich UI, real-time WebSocket, charts / Larger binary, separate build step
  3. **Separate deployment**: Simplest Go code / Violates single-binary requirement
- **Choice**: Embedded React SPA
- **Rationale**: Real-time query log with WebSocket streaming, interactive charts, and complex filtering require a proper SPA framework. React's ecosystem (Recharts, TanStack Query) provides production-ready solutions. Go's `embed` package makes embedding trivial.
- **Consequences**: Build pipeline needs Node.js for frontend compilation. Binary size increases by ~2-5MB (gzipped SPA assets). Frontend and backend versions are coupled.

#### Decision: In-memory radix tree for Gravity (blocklist)

- **Context**: Gravity must support sub-millisecond lookups across up to 5M domains (SPEC SS3.2, SS10.1).
- **Options Considered**:
  1. **HashMap**: O(1) lookup / High memory for 5M entries with full domain strings
  2. **Radix/Patricia tree**: Compressed trie, memory-efficient for domains / Slightly slower than hash
  3. **Bloom filter + hash**: Very low memory / False positives require secondary lookup
- **Choice**: Radix tree (hashicorp/go-immutable-radix or custom)
- **Rationale**: Domains share common suffixes (.com, .net, .org). Radix tree exploits this for ~60% memory reduction vs. hashmap. Wildcard matching (*.ads.example.com) is native to trie traversal. Immutable variant enables lock-free reads during Gravity rebuild.
- **Consequences**: Need custom domain reversal for suffix-based lookup (store "com.example.ads" for efficient wildcard matching).

#### Decision: Token bucket for rate limiting

- **Context**: API rate limiting at 60 req/min per session (SPEC SS6.4).
- **Choice**: Token bucket with in-memory store (golang.org/x/time/rate).
- **Rationale**: Stdlib-adjacent package, allows bursts (good UX), simple per-session tracking. No external dependency needed at Mantis's scale.

### 1.3 Dependency Inventory

**Philosophy:** Stdlib-first. External deps only when stdlib alternative is significantly worse or missing.

| Package                           | Purpose                         | License    | Justification                                                    |
|-----------------------------------|---------------------------------|------------|------------------------------------------------------------------|
| codeberg.org/miekg/dns            | DNS protocol library            | BSD-3      | No stdlib DNS server. De facto standard for Go DNS               |
| github.com/cockroachdb/pebble     | Key-value storage               | BSD-3      | Write-optimized LSM. Query log needs high write throughput       |
| github.com/go-chi/chi/v5          | HTTP router                     | MIT        | Middleware composition, stdlib compatible. Cleaner than raw mux  |
| golang.org/x/crypto               | ACME client, bcrypt, TLS        | BSD-3      | Official extended crypto. Let's Encrypt auto-cert                |
| golang.org/x/time/rate            | Rate limiter                    | BSD-3      | Token bucket. Stdlib-adjacent, well-tested                       |
| github.com/hashicorp/go-immutable-radix | Immutable radix tree      | MPL-2.0    | Lock-free concurrent reads for Gravity. Wildcard domain matching |
| github.com/pelletier/go-toml/v2   | TOML config parser              | MIT        | Config file parsing. Faster and stricter than BurntSushi/toml    |
| github.com/rs/zerolog             | Structured logging              | MIT        | Zero-allocation JSON logger. Performance-critical for DNS server |
| github.com/stretchr/testify       | Test assertions                 | MIT        | Readable test assertions, widely used                            |
| github.com/insomniacslk/dhcp      | DHCP protocol library           | BSD-3      | DHCPv4 server/client. Avoids reimplementing DHCP from scratch    |
| github.com/gorilla/websocket      | WebSocket connections           | BSD-3      | Real-time query log streaming. Mature, battle-tested             |
| github.com/google/uuid            | UUID generation                 | BSD-3      | Entity IDs. Faster than stdlib alternatives                      |

**Total direct dependencies: 12** (targeting <15 direct deps)

## 2. Design Patterns

### 2.1 Architectural Pattern: Clean Architecture (Ports & Adapters)

**Why:** DNS queries arrive via 4 different transports (UDP, TCP, DoH, DoT) but all share the same blocking/resolution logic (SPEC SS3.1, SS4.1). Core domain must be transport-agnostic.

**Application:** Domain layer defines interfaces (ports). Transport layers (DNS/HTTP/DHCP) are adapters. Pebble storage is an adapter. This allows testing blocking logic without any network or database.

**Code Sketch:**
```go
// internal/domain/resolver.go -- PORT (interface)
type Resolver interface {
    Resolve(ctx context.Context, query *Query) (*Response, error)
}

// internal/domain/query.go -- DOMAIN ENTITY
type Query struct {
    Domain    string
    Type      uint16
    ClientIP  net.IP
    Transport Transport // UDP, TCP, DOH, DOT
}

// internal/dns/server.go -- ADAPTER (transport)
type Server struct {
    resolver domain.Resolver
}

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
    query := domain.QueryFromDNS(r, w.RemoteAddr())
    resp, err := s.resolver.Resolve(context.Background(), query)
    // marshal resp back to DNS wire format
}
```

### 2.2 Chain of Responsibility: DNS Query Pipeline

**Why:** Each DNS query passes through multiple processing stages: cache lookup, blocklist check, client-specific rules, upstream resolution. Stages must be composable and orderable (SPEC SS3.1.1, SS3.2).

**Application:** Query handlers form a chain. Each handler either produces a response or passes to the next handler. New stages (e.g., conditional forwarding in v1.2) can be inserted without changing existing code.

**Code Sketch:**
```go
// internal/pipeline/handler.go
type QueryHandler interface {
    Handle(ctx context.Context, q *domain.Query, next QueryHandler) (*domain.Response, error)
}

// internal/pipeline/chain.go
type Chain struct {
    handlers []QueryHandler
}

func (c *Chain) Resolve(ctx context.Context, q *domain.Query) (*domain.Response, error) {
    return c.handlers[0].Handle(ctx, q, &chainLink{c.handlers, 1})
}

// Pipeline assembly:
// CacheHandler → GravityHandler → ClientRuleHandler → UpstreamHandler
```

### 2.3 Strategy Pattern: Resolver Backends

**Why:** Mantis supports two fundamentally different resolution strategies -- forwarding and recursive (SPEC SS3.1.4, SS3.1.5). Users switch between them via configuration.

**Application:** Both forwarding and recursive resolvers implement the same interface. The active strategy is selected at startup and can be switched via API without restart.

**Code Sketch:**
```go
// internal/resolver/forwarder.go
type Forwarder struct {
    upstreams []Upstream
    client    *dns.Client
}

func (f *Forwarder) Resolve(ctx context.Context, q *domain.Query) (*domain.Response, error) {
    // parallel query to all upstreams, return fastest
}

// internal/resolver/recursive.go
type Recursive struct {
    rootHints []dns.NS
    cache     *DNSCache
}

func (r *Recursive) Resolve(ctx context.Context, q *domain.Query) (*domain.Response, error) {
    // iterative resolution from root hints
}
```

### 2.4 Observer Pattern: Event Bus for Query Logging and Streaming

**Why:** Query processing must emit events consumed by multiple subscribers: log writer, statistics aggregator, WebSocket streamer (SPEC SS3.4.1, SS3.4.2, SS6.2 /queries/stream). Publishers and subscribers must be decoupled.

**Application:** Lightweight in-process event bus using Go channels. Query events are fanned out to all subscribers without blocking the DNS resolution path.

**Code Sketch:**
```go
// internal/event/bus.go
type Bus struct {
    subscribers []chan<- QueryEvent
    mu          sync.RWMutex
}

func (b *Bus) Publish(event QueryEvent) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    for _, sub := range b.subscribers {
        select {
        case sub <- event:
        default: // drop if subscriber is slow
        }
    }
}

func (b *Bus) Subscribe(bufSize int) <-chan QueryEvent {
    ch := make(chan QueryEvent, bufSize)
    b.mu.Lock()
    b.subscribers = append(b.subscribers, ch)
    b.mu.Unlock()
    return ch
}
```

### 2.5 Circuit Breaker: Upstream Health Management

**Why:** Upstream DNS servers can become unresponsive. Mantis must detect this and failover quickly rather than waiting for timeouts on every query (SPEC SS3.1.4).

**Application:** Each upstream server has a circuit breaker. After N consecutive failures, the circuit opens and queries skip that upstream. Periodic health checks attempt to close the circuit.

**Code Sketch:**
```go
// internal/resolver/circuit.go
type CircuitBreaker struct {
    state        State // Closed, Open, HalfOpen
    failures     int
    threshold    int
    lastFailure  time.Time
    cooldown     time.Duration
}

func (cb *CircuitBreaker) Allow() bool {
    switch cb.state {
    case Closed:
        return true
    case Open:
        if time.Since(cb.lastFailure) > cb.cooldown {
            cb.state = HalfOpen
            return true
        }
        return false
    case HalfOpen:
        return true // one test request
    }
    return false
}
```

### 2.6 Repository Pattern: Pebble Data Access

**Why:** Pebble is a low-level KV store. All modules need structured data access without knowing key encoding details (SPEC SS5.1). Repository pattern enables testing with in-memory fakes.

**Application:** Each entity type has a repository interface in the domain layer, with a Pebble implementation in the storage layer.

**Code Sketch:**
```go
// internal/domain/repository.go
type QueryLogRepository interface {
    Append(ctx context.Context, entry *QueryLogEntry) error
    Query(ctx context.Context, filter QueryLogFilter) ([]QueryLogEntry, error)
    DeleteBefore(ctx context.Context, before time.Time) (int64, error)
}

// internal/storage/pebble/querylog.go
type PebbleQueryLog struct {
    db *pebble.DB
}

// Key format: "qlog:<unix_nano_timestamp>:<sequence>"
func (r *PebbleQueryLog) Append(ctx context.Context, entry *QueryLogEntry) error {
    key := fmt.Appendf(nil, "qlog:%020d:%06d", entry.Timestamp.UnixNano(), seq)
    value, _ := entry.Marshal()
    return r.db.Set(key, value, pebble.Sync)
}
```

### 2.7 Middleware Pattern: HTTP API Pipeline

**Why:** All API endpoints share cross-cutting concerns: auth, rate limiting, CORS, logging, request ID (SPEC SS6.3, SS6.4).

**Application:** chi's built-in middleware composition. Each middleware is a standard http.Handler wrapper.

**Code Sketch:**
```go
// internal/api/router.go
func NewRouter(deps *Dependencies) chi.Router {
    r := chi.NewRouter()
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(zerolog.HTTPHandler(deps.Logger))
    r.Use(middleware.Recoverer)
    r.Use(corsMiddleware())

    r.Route("/api/v1", func(r chi.Router) {
        r.Post("/auth/login", deps.AuthHandler.Login)

        r.Group(func(r chi.Router) {
            r.Use(authMiddleware(deps.SessionStore))
            r.Use(rateLimitMiddleware(deps.RateLimiter))
            r.Get("/stats/summary", deps.StatsHandler.Summary)
            r.Get("/queries", deps.QueryHandler.List)
            // ... all authenticated endpoints
        })
    })
    return r
}
```

### 2.8 Worker Pool: Concurrent Blocklist Downloads

**Why:** Gravity rebuild downloads multiple blocklists concurrently (SPEC SS3.2.1). Need controlled parallelism to avoid overwhelming the network.

**Application:** Fixed-size worker pool for blocklist downloads. Results collected via channels and merged into the new Gravity tree.

**Code Sketch:**
```go
// internal/gravity/downloader.go
func (g *GravityEngine) downloadAll(ctx context.Context, sources []BlocklistSource) map[string][]string {
    results := make(chan downloadResult, len(sources))
    sem := make(chan struct{}, 5) // max 5 concurrent downloads

    for _, src := range sources {
        go func(s BlocklistSource) {
            sem <- struct{}{}
            defer func() { <-sem }()
            domains, err := g.download(ctx, s)
            results <- downloadResult{source: s, domains: domains, err: err}
        }(src)
    }
    // collect all results...
}
```

## 3. Project Structure

### 3.1 Directory Layout

```
mantis/
├── cmd/
│   └── mantis/
│       └── main.go                    # Entry point: flag parsing, config load, wire deps, start servers
├── internal/
│   ├── domain/                        # Core domain types and interfaces (no external deps)
│   │   ├── query.go                   # Query, Response, Transport types
│   │   ├── blocklist.go               # BlocklistSource, CustomRule entities
│   │   ├── lease.go                   # DhcpLease entity
│   │   ├── client.go                  # Client entity (network device)
│   │   ├── stats.go                   # Statistics aggregation types
│   │   ├── config.go                  # Configuration domain types
│   │   ├── repository.go             # All repository interfaces
│   │   └── errors.go                  # Domain error types
│   ├── dns/                           # DNS transport layer (UDP/TCP server)
│   │   ├── server.go                  # DNS server using miekg/dns
│   │   ├── handler.go                 # DNS message → domain.Query adapter
│   │   └── writer.go                  # domain.Response → DNS message adapter
│   ├── doh/                           # DNS-over-HTTPS transport
│   │   ├── server.go                  # HTTPS server for DoH
│   │   └── handler.go                 # HTTP ↔ DNS wire format
│   ├── dot/                           # DNS-over-TLS transport
│   │   ├── server.go                  # TLS listener for DoT
│   │   └── handler.go                 # TLS ↔ DNS wire format
│   ├── dhcp/                          # DHCP server
│   │   ├── server.go                  # DHCPv4 server
│   │   ├── handler.go                 # DHCP message handling
│   │   ├── pool.go                    # IP address pool management
│   │   └── lease.go                   # Lease lifecycle management
│   ├── pipeline/                      # DNS query processing chain
│   │   ├── chain.go                   # Chain of responsibility orchestrator
│   │   ├── cache.go                   # Cache lookup/store handler
│   │   ├── gravity.go                 # Blocklist lookup handler
│   │   ├── clientrule.go             # Per-client blocking handler
│   │   └── upstream.go                # Forward/resolve handler
│   ├── resolver/                      # DNS resolution backends
│   │   ├── forwarder.go               # Upstream forwarding (Strategy A)
│   │   ├── recursive.go               # Recursive resolution (Strategy B)
│   │   ├── circuit.go                 # Circuit breaker for upstreams
│   │   └── cache.go                   # DNS response cache (TTL-aware)
│   ├── gravity/                       # Blocklist management engine
│   │   ├── engine.go                  # Gravity rebuild orchestrator
│   │   ├── downloader.go             # HTTP blocklist fetcher
│   │   ├── parser.go                  # Hosts-file, domain-list, adblock parsers
│   │   ├── tree.go                    # Radix tree wrapper for domain lookup
│   │   └── scheduler.go              # Periodic rebuild scheduler
│   ├── event/                         # Event bus for query events
│   │   ├── bus.go                     # Publish/subscribe event bus
│   │   └── types.go                   # Event type definitions
│   ├── api/                           # REST API layer
│   │   ├── router.go                  # chi router setup, middleware chain
│   │   ├── auth.go                    # Login, session, API key handlers
│   │   ├── stats.go                   # Statistics endpoint handlers
│   │   ├── queries.go                 # Query log endpoint + WebSocket
│   │   ├── blocklists.go             # Blocklist CRUD handlers
│   │   ├── rules.go                   # Custom rule CRUD handlers
│   │   ├── dhcp.go                    # DHCP management handlers
│   │   ├── clients.go                # Client management handlers
│   │   ├── settings.go               # Settings CRUD handlers
│   │   ├── system.go                  # System info, health handlers
│   │   ├── middleware.go              # Auth, rate limit, CORS middleware
│   │   └── response.go               # Standard response/error helpers
│   ├── storage/                       # Pebble storage implementations
│   │   ├── pebble.go                  # Pebble DB lifecycle (open, close, compact)
│   │   ├── querylog.go               # QueryLogRepository implementation
│   │   ├── blocklist.go              # BlocklistRepository implementation
│   │   ├── rules.go                   # CustomRuleRepository implementation
│   │   ├── leases.go                 # LeaseRepository implementation
│   │   ├── settings.go               # SettingsRepository implementation
│   │   ├── sessions.go               # SessionRepository implementation
│   │   └── keys.go                    # Key encoding/decoding helpers
│   ├── stats/                         # Statistics aggregation engine
│   │   ├── aggregator.go             # Time-bucketed query aggregation
│   │   └── collector.go              # Listens to event bus, updates aggregates
│   ├── tls/                           # TLS certificate management
│   │   ├── manager.go                # Certificate loading + ACME orchestration
│   │   └── acme.go                   # Let's Encrypt ACME client wrapper
│   ├── config/                        # Configuration management
│   │   ├── config.go                  # Config struct + defaults
│   │   ├── loader.go                 # TOML file + env + flags merge
│   │   └── validate.go               # Config validation rules
│   └── web/                           # Embedded frontend
│       └── embed.go                   # go:embed directive for SPA assets
├── web/                               # React SPA source (not embedded, built separately)
│   ├── src/
│   │   ├── main.tsx                   # React entry point
│   │   ├── App.tsx                    # Root component, router setup
│   │   ├── api/                       # API client layer
│   │   │   ├── client.ts             # Axios/fetch wrapper with auth
│   │   │   ├── queries.ts            # TanStack Query hooks for query log
│   │   │   ├── stats.ts              # TanStack Query hooks for statistics
│   │   │   ├── blocklists.ts         # TanStack Query hooks for blocklists
│   │   │   ├── rules.ts              # TanStack Query hooks for rules
│   │   │   ├── dhcp.ts               # TanStack Query hooks for DHCP
│   │   │   ├── settings.ts           # TanStack Query hooks for settings
│   │   │   └── websocket.ts          # WebSocket client for live queries
│   │   ├── components/                # Shared UI components
│   │   │   ├── Layout.tsx            # Sidebar + content layout
│   │   │   ├── Sidebar.tsx           # Navigation sidebar
│   │   │   ├── DataTable.tsx         # Generic sortable/filterable table
│   │   │   ├── StatCard.tsx          # Dashboard summary card
│   │   │   ├── Chart.tsx             # Recharts wrapper components
│   │   │   ├── SearchInput.tsx       # Search bar with debounce
│   │   │   ├── Toggle.tsx            # On/off toggle switch
│   │   │   ├── Modal.tsx             # Modal dialog
│   │   │   ├── Toast.tsx             # Notification toast
│   │   │   └── ThemeToggle.tsx       # Dark/light mode switch
│   │   ├── pages/                     # Route-level page components
│   │   │   ├── Dashboard.tsx         # Main dashboard (SPEC SS7.2.1)
│   │   │   ├── QueryLog.tsx          # Query log with live stream (SPEC SS7.2.2)
│   │   │   ├── Blocklists.tsx        # Blocklist management (SPEC SS7.2.3)
│   │   │   ├── Rules.tsx             # Custom allow/block rules (SPEC SS7.2.4)
│   │   │   ├── Dhcp.tsx              # DHCP management (SPEC SS7.2.5)
│   │   │   ├── Network.tsx           # Client devices (SPEC SS7.2.6)
│   │   │   ├── Settings.tsx          # System settings (SPEC SS7.2.7)
│   │   │   ├── Statistics.tsx        # Long-term stats (SPEC SS7.2.8)
│   │   │   └── Login.tsx             # Authentication page
│   │   ├── store/                     # Zustand client-side stores
│   │   │   ├── authStore.ts          # Auth session state
│   │   │   └── uiStore.ts           # Theme, sidebar, UI preferences
│   │   ├── hooks/                     # Custom React hooks
│   │   │   ├── useWebSocket.ts       # WebSocket connection hook
│   │   │   └── useDebounce.ts        # Debounce hook for search
│   │   └── utils/                     # Utility functions
│   │       ├── format.ts             # Number, date, domain formatting
│   │       └── constants.ts          # API URLs, default values
│   ├── index.html
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── package.json
│   └── .eslintrc.cjs
├── configs/
│   └── mantis.example.toml           # Example configuration file
├── scripts/
│   ├── install.sh                     # Linux install script (systemd)
│   └── build.sh                       # Cross-compilation build script
├── docker/
│   ├── Dockerfile                     # Multi-stage production build
│   └── docker-compose.yml            # Example compose with recommended settings
├── docs/
│   ├── SPECIFICATION.md
│   ├── IMPLEMENTATION.md
│   ├── TASKS.md
│   └── BRANDING.md
├── go.mod
├── go.sum
├── Makefile
├── LICENSE
└── README.md
```

**Structural Philosophy:**
- **Layer-based with domain at center**: `internal/domain/` has zero external imports. All other packages depend on domain, not each other.
- **Transport separation**: DNS, DoH, DoT, DHCP, and HTTP API each have their own package. All convert to/from domain types.
- **Internal boundary**: Everything under `internal/` is private to the module. Only `cmd/mantis/main.go` wires it together.
- **Frontend co-located**: `web/` directory at root for SPA source. Built artifacts embedded via `internal/web/embed.go`.
- **Tests co-located**: `_test.go` files next to source. Integration tests in `tests/` directory if needed.

### 3.2 Module Breakdown

#### Module: domain

- **Path**: `internal/domain/`
- **Responsibility**: Define all core types, entities, and interfaces. Zero external dependencies.
- **Exports**: Query, Response, BlocklistSource, CustomRule, QueryLogEntry, DhcpLease, all repository interfaces, domain errors
- **Imports**: Only stdlib (net, time, errors)
- **Key Files**:
  - `query.go` -- Query and Response types, Transport enum
  - `repository.go` -- All repository interfaces (QueryLogRepository, BlocklistRepository, etc.)
  - `errors.go` -- Typed errors (ErrBlocked, ErrUpstreamTimeout, ErrNotFound, etc.)

#### Module: dns

- **Path**: `internal/dns/`
- **Responsibility**: UDP/TCP DNS server. Translates between miekg/dns messages and domain types.
- **Exports**: Server struct with Start/Stop methods
- **Imports**: domain, miekg/dns
- **Key Files**:
  - `server.go` -- DNS server lifecycle (bind, listen, shutdown)
  - `handler.go` -- Converts dns.Msg to domain.Query, calls resolver pipeline

#### Module: pipeline

- **Path**: `internal/pipeline/`
- **Responsibility**: DNS query processing chain. Orchestrates cache, blocking, resolution.
- **Exports**: Chain struct, QueryHandler interface
- **Imports**: domain, resolver, gravity, event
- **Key Files**:
  - `chain.go` -- Chain of responsibility orchestrator
  - `gravity.go` -- Checks domain against Gravity blocklist
  - `cache.go` -- DNS response cache handler
  - `upstream.go` -- Delegates to resolver (forward or recursive)

#### Module: resolver

- **Path**: `internal/resolver/`
- **Responsibility**: DNS resolution backends -- forwarding and recursive.
- **Exports**: Forwarder, Recursive (both implement domain.Resolver), CircuitBreaker
- **Imports**: domain, miekg/dns
- **Key Files**:
  - `forwarder.go` -- Parallel upstream query with failover
  - `recursive.go` -- Iterative resolution from root hints
  - `circuit.go` -- Circuit breaker state machine per upstream
  - `cache.go` -- TTL-aware DNS response cache

#### Module: gravity

- **Path**: `internal/gravity/`
- **Responsibility**: Blocklist lifecycle -- download, parse, deduplicate, compile to radix tree.
- **Exports**: Engine struct (Rebuild, Lookup, AddRule, RemoveRule methods)
- **Imports**: domain, hashicorp/go-immutable-radix
- **Key Files**:
  - `engine.go` -- Gravity rebuild orchestrator, atomic tree swap
  - `parser.go` -- Multi-format blocklist parser (hosts, domains, adblock)
  - `tree.go` -- Radix tree with reversed-domain keys for wildcard matching

#### Module: dhcp

- **Path**: `internal/dhcp/`
- **Responsibility**: DHCPv4 server with IP pool management and lease persistence.
- **Exports**: Server struct with Start/Stop
- **Imports**: domain, insomniacslk/dhcp
- **Key Files**:
  - `server.go` -- DHCP server lifecycle
  - `pool.go` -- IP range management, allocation, exhaustion detection
  - `lease.go` -- Lease creation, renewal, expiry, static reservations

#### Module: api

- **Path**: `internal/api/`
- **Responsibility**: REST API + WebSocket for admin UI and automation.
- **Exports**: Router constructor
- **Imports**: domain, chi, gorilla/websocket, zerolog
- **Key Files**:
  - `router.go` -- Route definitions, middleware chain assembly
  - `queries.go` -- Query log list + WebSocket live stream
  - `auth.go` -- Login, logout, session validation, API key management

#### Module: storage

- **Path**: `internal/storage/`
- **Responsibility**: Pebble implementations of all repository interfaces.
- **Exports**: All repository implementations, NewDB constructor
- **Imports**: domain, pebble
- **Key Files**:
  - `pebble.go` -- DB lifecycle, compaction, backup
  - `keys.go` -- Key encoding (prefix + timestamp + ID patterns)
  - `querylog.go` -- High-throughput query log writes with batching

#### Module: event

- **Path**: `internal/event/`
- **Responsibility**: In-process pub/sub event bus.
- **Exports**: Bus struct, QueryEvent type
- **Imports**: Only stdlib
- **Key Files**:
  - `bus.go` -- Publish, Subscribe, Unsubscribe with channel-based delivery

#### Module: config

- **Path**: `internal/config/`
- **Responsibility**: Load configuration from TOML file, env vars, and CLI flags.
- **Exports**: Config struct, Load function
- **Imports**: go-toml
- **Key Files**:
  - `config.go` -- Config struct with all fields and defaults
  - `loader.go` -- Merge priority: flags > env > file > defaults
  - `validate.go` -- Validation rules (port ranges, IP formats, etc.)

#### Module: tls

- **Path**: `internal/tls/`
- **Responsibility**: TLS certificate management and ACME auto-provisioning.
- **Exports**: Manager struct with GetCertificate, certificates for DoH/DoT
- **Imports**: golang.org/x/crypto/acme/autocert
- **Key Files**:
  - `manager.go` -- Certificate loading, renewal, fallback logic
  - `acme.go` -- Let's Encrypt ACME client wrapper

### 3.3 Module Dependency Graph

```
cmd/mantis/main.go
    │
    ├──→ config
    │
    ├──→ storage ──→ domain
    │       │
    │       └──→ pebble (external)
    │
    ├──→ gravity ──→ domain
    │       │
    │       └──→ go-immutable-radix (external)
    │
    ├──→ pipeline ──→ domain
    │       │
    │       ├──→ gravity
    │       ├──→ resolver
    │       └──→ event
    │
    ├──→ resolver ──→ domain
    │       │
    │       └──→ miekg/dns (external)
    │
    ├──→ dns ──→ domain, pipeline
    │       │
    │       └──→ miekg/dns (external)
    │
    ├──→ doh ──→ domain, pipeline, tls
    │
    ├──→ dot ──→ domain, pipeline, tls
    │
    ├──→ dhcp ──→ domain, storage
    │       │
    │       └──→ insomniacslk/dhcp (external)
    │
    ├──→ api ──→ domain, storage, gravity, event, stats
    │       │
    │       ├──→ chi (external)
    │       └──→ gorilla/websocket (external)
    │
    ├──→ stats ──→ domain, event, storage
    │
    ├──→ tls
    │       │
    │       └──→ x/crypto (external)
    │
    ├──→ event
    │
    └──→ web (embedded SPA)
```

**No circular dependencies.** domain is the leaf -- everything depends on it, it depends on nothing.

## 4. Data Layer

### 4.1 Pebble Key Schema

Pebble is a key-value store. Key design is critical for efficient range scans.

**Key prefix convention:** `<entity>:<sort_key>` where keys are lexicographically ordered.

```
Prefix          | Key Format                                    | Value
----------------|-----------------------------------------------|---------------------------
qlog:           | qlog:<unix_nano_20digits>:<seq_6digits>       | QueryLogEntry (protobuf)
blsrc:          | blsrc:<uuid>                                  | BlocklistSource (JSON)
rule:           | rule:<uuid>                                   | CustomRule (JSON)
lease:          | lease:<mac_normalized>                         | DhcpLease (JSON)
lease-ip:       | lease-ip:<ipv4_padded>                        | MAC address (cross-index)
setting:        | setting:<dotted.key.path>                     | JSON value
session:        | session:<token_hash>                          | Session (JSON)
apikey:         | apikey:<key_hash>                             | APIKey (JSON)
stat:           | stat:<bucket_type>:<timestamp>                | StatBucket (protobuf)
gravity-meta:   | gravity-meta:last-rebuild                     | Timestamp
```

**Key design rationale:**
- `qlog:` prefix enables time-range scans (PrefixIterator + SeekGE for start time). Zero-padded nanosecond timestamp ensures lexicographic = chronological order.
- `lease:` keyed by MAC for O(1) lookup on DHCP REQUEST. `lease-ip:` cross-index for reverse lookup.
- `stat:` bucketed by hour/day for efficient dashboard queries without full log scans.

### 4.2 Serialization Strategy

- **QueryLogEntry and StatBucket**: Protocol Buffers (high volume, needs compact encoding)
- **All other entities**: JSON (human-readable, easy debugging, low volume)

Proto definition:
```protobuf
message QueryLogEntry {
    fixed64 timestamp_ns = 1;
    bytes client_ip = 2;      // 4 or 16 bytes
    string domain = 3;
    uint32 query_type = 4;
    Result result = 5;
    string upstream = 6;
    int64 latency_us = 7;
    string answer = 8;
}
```

### 4.3 Data Access Pattern

**Write batching for query log:** Individual Pebble writes per query would bottleneck at ~50K ops/sec. Solution: batch writes.

```go
type BatchWriter struct {
    db       *pebble.DB
    batch    *pebble.Batch
    count    int
    maxBatch int           // flush every N entries (default 1000)
    maxAge   time.Duration // flush every N ms (default 100ms)
    ticker   *time.Ticker
    mu       sync.Mutex
}

func (w *BatchWriter) Append(key, value []byte) error {
    w.mu.Lock()
    defer w.mu.Unlock()
    w.batch.Set(key, value, nil)
    w.count++
    if w.count >= w.maxBatch {
        return w.flush()
    }
    return nil
}
```

### 4.4 Caching Strategy

**DNS Cache (in-memory, not Pebble):**
- LRU cache with TTL-aware eviction
- Maximum entries: configurable (default 10,000)
- Key: domain + query type
- Value: DNS response + insertion time + original TTL
- Eviction: min(TTL expiry, LRU when full)
- Thread-safe via sync.RWMutex (read-heavy workload)

**Statistics Cache:**
- Current-hour bucket kept in memory, flushed to Pebble on hour rollover
- Dashboard summary computed from last 24 hourly buckets (24 Pebble reads)
- No cache invalidation needed -- buckets are immutable once the hour passes

## 5. API Implementation

### 5.1 Route Structure

| Method | Path                            | Handler              | Middleware       | Spec Ref |
|--------|---------------------------------|----------------------|------------------|----------|
| POST   | /api/v1/auth/login              | auth.Login           | rateLimit(5/min) | SS6.3    |
| POST   | /api/v1/auth/logout             | auth.Logout          | auth             | SS6.3    |
| GET    | /api/v1/stats/summary           | stats.Summary        | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/stats/overtime           | stats.OverTime       | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/stats/top-domains       | stats.TopDomains     | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/stats/top-clients       | stats.TopClients     | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/queries                 | queries.List         | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/queries/stream          | queries.Stream (WS)  | auth             | SS6.2    |
| GET    | /api/v1/blocklists              | blocklists.List      | auth, rateLimit  | SS6.2    |
| POST   | /api/v1/blocklists              | blocklists.Create    | auth, rateLimit  | SS6.2    |
| PUT    | /api/v1/blocklists/:id          | blocklists.Update    | auth, rateLimit  | SS6.2    |
| DELETE | /api/v1/blocklists/:id          | blocklists.Delete    | auth, rateLimit  | SS6.2    |
| POST   | /api/v1/gravity/rebuild         | gravity.Rebuild      | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/gravity/status          | gravity.Status       | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/rules                   | rules.List           | auth, rateLimit  | SS6.2    |
| POST   | /api/v1/rules                   | rules.Create         | auth, rateLimit  | SS6.2    |
| DELETE | /api/v1/rules/:id               | rules.Delete         | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/dhcp/leases             | dhcp.ListLeases      | auth, rateLimit  | SS6.2    |
| POST   | /api/v1/dhcp/leases/static      | dhcp.CreateStatic    | auth, rateLimit  | SS6.2    |
| DELETE | /api/v1/dhcp/leases/static/:mac | dhcp.DeleteStatic    | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/dhcp/config             | dhcp.GetConfig       | auth, rateLimit  | SS6.2    |
| PUT    | /api/v1/dhcp/config             | dhcp.UpdateConfig    | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/clients                 | clients.List         | auth, rateLimit  | SS6.2    |
| PUT    | /api/v1/clients/:ip/blocking    | clients.SetBlocking  | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/settings                | settings.Get         | auth, rateLimit  | SS6.2    |
| PUT    | /api/v1/settings                | settings.Update      | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/system/info             | system.Info          | auth, rateLimit  | SS6.2    |
| POST   | /api/v1/system/restart-dns      | system.RestartDNS    | auth, rateLimit  | SS6.2    |
| GET    | /api/v1/dns/test                | dns.Test             | auth, rateLimit  | SS6.2    |

### 5.2 Request/Response Contract

**Success Response:**
```json
{
  "data": { ... },
  "meta": {
    "requestId": "req_abc123",
    "page": 1,
    "perPage": 50,
    "total": 1234
  }
}
```

**Error Response (matches SPEC SS6.5):**
```json
{
  "error": {
    "code": "BLOCKLIST_DOWNLOAD_FAILED",
    "message": "Failed to download blocklist: connection timeout",
    "requestId": "req_abc123"
  }
}
```

### 5.3 Validation Approach

Struct-tag validation via `go-playground/validator` or manual validation in handler layer. Domain names validated against RFC 1035. IP addresses via `net.ParseIP`.

```go
type CreateBlocklistRequest struct {
    Name    string `json:"name" validate:"required,min=1,max=256"`
    URL     string `json:"url" validate:"required,url"`
    Format  string `json:"format" validate:"required,oneof=hosts domains adblock"`
    Enabled *bool  `json:"enabled"`
}
```

### 5.4 Authentication Flow

```
1. First run: Setup wizard prompts for admin password → stored bcrypt-hashed in Pebble
2. Login: POST /auth/login with {username, password}
3. Server validates bcrypt → creates session → returns Set-Cookie (HTTP-only, Secure, SameSite=Strict)
4. Subsequent requests: Cookie sent automatically → authMiddleware validates session in Pebble
5. API key auth: X-Api-Key header → middleware looks up hashed key in Pebble
6. Session expiry: 24h sliding window. Each valid request extends expiry.
7. Logout: POST /auth/logout → deletes session from Pebble
```

## 6. Frontend Implementation

### 6.1 Component Architecture

```
App
├── Layout (sidebar + content area)
│   ├── Sidebar (navigation + theme toggle)
│   └── <Routes>
│       ├── Dashboard
│       │   ├── StatCard (x4: total queries, blocked %, cache hit, clients)
│       │   ├── QueryChart (24h line chart)
│       │   ├── TopBlockedDomains (table)
│       │   └── TopAllowedDomains (table)
│       ├── QueryLog
│       │   ├── SearchInput + FilterBar
│       │   ├── LiveToggle (WebSocket on/off)
│       │   └── DataTable (paginated, sortable)
│       ├── Blocklists
│       │   ├── BlocklistTable (toggle, stats, delete)
│       │   └── AddBlocklistModal
│       ├── Rules
│       │   ├── RulesTable (allow/block tabs)
│       │   └── AddRuleModal
│       ├── Dhcp
│       │   ├── ActiveLeases (table)
│       │   ├── StaticLeases (table + add form)
│       │   └── PoolConfig (form)
│       ├── Network
│       │   ├── ClientTable (hostname, IP, MAC, queries, blocking toggle)
│       │   └── ClientDetail (modal with query history)
│       ├── Settings
│       │   └── TabPanel (DNS, DHCP, Encryption, Logging, System)
│       └── Statistics
│           ├── DateRangePicker
│           └── Charts (queries/day, block rate, top domains)
└── Login (unauthenticated route)
```

### 6.2 State Management

- **Server state**: TanStack Query v5. All API data fetched/cached/invalidated via query hooks. Stale-while-revalidate pattern for dashboard freshness.
- **Client state**: Zustand. Two stores:
  - `authStore`: session token, user info, isAuthenticated
  - `uiStore`: theme (dark/light), sidebar collapsed, table page sizes
- **WebSocket state**: Custom `useWebSocket` hook managing connection lifecycle, auto-reconnect, and query event stream buffer.

### 6.3 Routing

React Router v7 with layout routes:

```typescript
const router = createBrowserRouter([
  { path: "/login", element: <Login /> },
  {
    path: "/",
    element: <AuthGuard><Layout /></AuthGuard>,
    children: [
      { index: true, element: <Dashboard /> },
      { path: "queries", element: <QueryLog /> },
      { path: "blocklists", element: <Blocklists /> },
      { path: "rules", element: <Rules /> },
      { path: "dhcp", element: <Dhcp /> },
      { path: "network", element: <Network /> },
      { path: "settings", element: <Settings /> },
      { path: "stats", element: <Statistics /> },
    ],
  },
]);
```

### 6.4 Styling

- **Tailwind CSS v4**: Utility-first, tree-shaken in production. Dark mode via `class` strategy.
- **Design tokens**: CSS custom properties for colors, spacing, radii. Defined in `tailwind.config.ts`.
- **Responsive**: Desktop-first. Sidebar collapses at 1024px. Simplified layout at 768px.
- **Component library**: No external UI library. Custom components with Tailwind for minimal bundle.

## 7. Error Handling Strategy

### 7.1 Error Classification

| Category   | Example                     | DNS Response | HTTP Code | Logged As | User Sees               |
|------------|-----------------------------|--------------|-----------|-----------|-----------------------  |
| Blocked    | Domain on blocklist         | 0.0.0.0      | N/A       | Info      | N/A (DNS client)        |
| Cached     | Response in cache           | Cached data  | N/A       | Debug     | N/A (DNS client)        |
| Upstream   | Upstream timeout            | SERVFAIL     | N/A       | Warn      | N/A (DNS client)        |
| Validation | Bad API input               | N/A          | 400       | Debug     | Field error             |
| Auth       | Invalid session             | N/A          | 401       | Info      | "Session expired"       |
| Not Found  | Unknown blocklist ID        | N/A          | 404       | Debug     | "Not found"             |
| Rate Limit | Too many requests           | N/A          | 429       | Info      | "Rate limit exceeded"   |
| Internal   | Pebble write failure        | SERVFAIL     | 500       | Error     | "Internal error"        |

### 7.2 Error Propagation

```
Domain error (internal/domain/errors.go)
    → Returned by repository/service layer
    → Handler maps to HTTP status via errors.As/errors.Is
    → API response serialized via response.Error()
    → Logged at appropriate level by middleware
```

DNS errors follow a separate path:
```
Resolution error → Pipeline catches → Returns SERVFAIL or cached response → Logs via event bus
```

## 8. Configuration

### 8.1 Config Sources

Priority (highest wins): CLI flags → Environment variables → Config file → Defaults

### 8.2 Config Schema

| Key                         | Type     | Default                    | Env Var                      | Description                          |
|-----------------------------|----------|----------------------------|------------------------------|--------------------------------------|
| dns.listenAddress           | string   | "0.0.0.0:53"              | MANTIS_DNS_LISTEN            | DNS server bind address:port         |
| dns.upstreams               | []string | ["1.1.1.1", "8.8.8.8"]   | MANTIS_DNS_UPSTREAMS         | Upstream DNS servers                 |
| dns.resolverMode            | string   | "forward"                  | MANTIS_DNS_RESOLVER_MODE     | "forward" or "recursive"            |
| dns.cacheSize               | int      | 10000                      | MANTIS_DNS_CACHE_SIZE        | Max DNS cache entries                |
| dns.blockingMode            | string   | "null"                     | MANTIS_DNS_BLOCKING_MODE     | "null" (0.0.0.0) or "nxdomain"     |
| doh.enabled                 | bool     | false                      | MANTIS_DOH_ENABLED           | Enable DNS-over-HTTPS                |
| doh.listenAddress           | string   | "0.0.0.0:443"             | MANTIS_DOH_LISTEN            | DoH bind address:port                |
| dot.enabled                 | bool     | false                      | MANTIS_DOT_ENABLED           | Enable DNS-over-TLS                  |
| dot.listenAddress           | string   | "0.0.0.0:853"             | MANTIS_DOT_LISTEN            | DoT bind address:port                |
| tls.certFile                | string   | ""                         | MANTIS_TLS_CERT              | TLS certificate file path            |
| tls.keyFile                 | string   | ""                         | MANTIS_TLS_KEY               | TLS private key file path            |
| tls.acmeEnabled             | bool     | false                      | MANTIS_TLS_ACME_ENABLED      | Enable Let's Encrypt auto-cert       |
| tls.acmeDomain              | string   | ""                         | MANTIS_TLS_ACME_DOMAIN       | Domain for ACME certificate          |
| tls.acmeEmail               | string   | ""                         | MANTIS_TLS_ACME_EMAIL        | Email for ACME registration          |
| dhcp.enabled                | bool     | false                      | MANTIS_DHCP_ENABLED          | Enable DHCP server                   |
| dhcp.interface              | string   | ""                         | MANTIS_DHCP_INTERFACE        | Network interface for DHCP           |
| dhcp.rangeStart             | string   | ""                         | MANTIS_DHCP_RANGE_START      | DHCP pool start IP                   |
| dhcp.rangeEnd               | string   | ""                         | MANTIS_DHCP_RANGE_END        | DHCP pool end IP                     |
| dhcp.leaseDuration          | duration | "24h"                      | MANTIS_DHCP_LEASE_DURATION   | Default lease duration               |
| dhcp.gateway                | string   | ""                         | MANTIS_DHCP_GATEWAY          | Default gateway IP                   |
| dhcp.subnetMask             | string   | "255.255.255.0"            | MANTIS_DHCP_SUBNET           | Subnet mask                          |
| api.listenAddress           | string   | "0.0.0.0:8080"            | MANTIS_API_LISTEN            | Admin UI/API bind address:port       |
| api.rateLimit               | int      | 60                         | MANTIS_API_RATE_LIMIT        | Requests per minute per session      |
| storage.dataDir             | string   | "/var/lib/mantis"          | MANTIS_DATA_DIR              | Pebble database directory            |
| logging.level               | string   | "info"                     | MANTIS_LOG_LEVEL             | Log level (debug/info/warn/error)    |
| logging.queryLog            | string   | "all"                      | MANTIS_QUERY_LOG             | "all", "blocked", "none"            |
| logging.retentionDays       | int      | 30                         | MANTIS_LOG_RETENTION         | Query log retention in days          |
| logging.privacyMode         | bool     | false                      | MANTIS_PRIVACY_MODE          | Anonymize client IPs in logs         |
| gravity.updateInterval      | duration | "24h"                      | MANTIS_GRAVITY_INTERVAL      | Blocklist update interval            |

## 9. Testing Strategy

### 9.1 Test Pyramid

| Level       | Tool                      | Scope                              | Target                              |
|-------------|---------------------------|------------------------------------|-------------------------------------|
| Unit        | Go testing + testify      | Functions, methods, single package | 80%+ on domain, pipeline, gravity   |
| Integration | Go testing + real Pebble  | Storage + domain together          | All repository implementations      |
| API         | Go httptest + chi         | Full HTTP request/response cycle   | All API endpoints                   |
| DNS         | miekg/dns client in tests | DNS query/response cycle           | Blocking, caching, forwarding       |
| Frontend    | Vitest + Testing Library  | React components                   | All pages and interactive components|
| E2E         | Playwright                | Browser + real backend             | Critical flows (login, block domain)|

### 9.2 Test Patterns

**Table-driven tests** for DNS resolution:
```go
func TestGravityLookup(t *testing.T) {
    tests := []struct {
        name     string
        domain   string
        blocked  bool
    }{
        {"exact match", "ads.example.com", true},
        {"wildcard match", "sub.ads.example.com", true},
        {"not blocked", "example.com", false},
        {"allowlisted", "allowed.ads.example.com", false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { ... })
    }
}
```

**In-memory Pebble** for storage tests:
```go
func newTestDB(t *testing.T) *pebble.DB {
    db, err := pebble.Open(t.TempDir(), &pebble.Options{})
    require.NoError(t, err)
    t.Cleanup(func() { db.Close() })
    return db
}
```

### 9.3 CI Pipeline

```
Push/PR → golangci-lint → go test ./... → go build (all targets) → npm run lint → npm test → npm run build → Docker build → Release (tags only)
```

Matrix: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64

## 10. Security Implementation

### 10.1 Input Sanitization Points

- **DNS queries**: Domain validated per RFC 1035 (max 253 chars, labels max 63, alphanumeric + hyphen)
- **API requests**: JSON body size limited to 1MB. All string fields trimmed and length-checked.
- **TOML config**: Strict parsing with go-toml/v2. Unknown keys cause error, not ignored.
- **Blocklist URLs**: Must be HTTP/HTTPS. No file:// or other schemes. Timeout on download.

### 10.2 Secret Management

- **Development**: `.env` file (gitignored) + `MANTIS_*` env vars
- **Production**: Environment variables or TOML config with restricted file permissions (0600)
- **Admin password**: bcrypt hashed (cost 12), stored in Pebble
- **Session tokens**: crypto/rand 256-bit, stored hashed (SHA-256) in Pebble
- **API keys**: crypto/rand 512-bit, stored hashed (SHA-256) in Pebble
- **TLS private key**: File system with 0600 permissions. Never logged or exposed via API.

### 10.3 Security Headers

```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "0")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Content-Security-Policy",
            "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
        next.ServeHTTP(w, r)
    })
}
```

## 11. Deployment

### 11.1 Build Commands

```bash
# Backend build (all platforms)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=${VERSION}" -o dist/mantis-linux-amd64 ./cmd/mantis
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=${VERSION}" -o dist/mantis-linux-arm64 ./cmd/mantis
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=${VERSION}" -o dist/mantis-darwin-amd64 ./cmd/mantis
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=${VERSION}" -o dist/mantis-darwin-arm64 ./cmd/mantis

# Frontend build (runs before Go build)
cd web && npm ci && npm run build && cd ..
# Built assets land in web/dist/, embedded via internal/web/embed.go
```

### 11.2 Dockerfile

```dockerfile
# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS backend
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /mantis ./cmd/mantis

# Stage 3: Minimal runtime
FROM scratch
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /mantis /mantis
EXPOSE 53/udp 53/tcp 67/udp 68/udp 443 853 8080
VOLUME /data
ENTRYPOINT ["/mantis"]
CMD ["--config", "/data/config.toml", "--data-dir", "/data"]
```

### 11.3 Health Check

```
GET /api/v1/system/health → 200 OK
```

```json
{
  "status": "healthy",
  "components": {
    "dns": "up",
    "storage": "up",
    "gravity": "ready",
    "dhcp": "up"
  },
  "version": "1.0.0",
  "uptime": "72h15m"
}
```

### 11.4 Monitoring

- **Logging**: zerolog structured JSON to stdout. Docker/systemd collect automatically.
- **Health endpoint**: `/api/v1/system/health` for Docker HEALTHCHECK and load balancers.
- **System info**: `/api/v1/system/info` returns memory, CPU, goroutine count, Pebble stats.
- **Prometheus**: Planned for v1.2. Not in v1.0 to limit scope.

## 12. Development Workflow

### 12.1 Local Setup

```bash
# Prerequisites: Go 1.26+, Node.js 22+
git clone https://github.com/[org]/mantis.git
cd mantis

# Backend
go mod download

# Frontend
cd web && npm ci && cd ..

# Run in development (frontend dev server + Go backend)
# Terminal 1: Frontend with hot reload
cd web && npm run dev

# Terminal 2: Go backend (recompile on change with air or manual)
go run ./cmd/mantis --config configs/mantis.example.toml --data-dir ./data

# Or build everything and run
make build
./dist/mantis --config configs/mantis.example.toml
```

### 12.2 Code Standards

- **Go**: golangci-lint with default + exhaustive, errcheck, gocritic, gosec linters
- **TypeScript**: ESLint + Prettier. Strict TypeScript (strict: true)
- **Commit convention**: Conventional Commits (feat:, fix:, docs:, refactor:, test:, chore:)
- **Pre-commit**: golangci-lint + npm run lint via Makefile target

### 12.3 Git Workflow

- **Branch naming**: `feat/dns-cache`, `fix/gravity-parse-error`, `refactor/storage-keys`
- **PR process**: Feature branch → PR → CI passes → review → squash merge to main
- **Releases**: Tag-based. Push `v1.0.0` tag → CI builds all platforms → GitHub Release with binaries + Docker image
