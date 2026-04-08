# Mantis -- Specification

> A high-performance, Go-native DNS sinkhole with DHCP, encrypted DNS, and a modern admin interface.

## 1. Overview

### 1.1 What Is Mantis?

Mantis is a network-level advertisement and tracker blocker written entirely in Go. It operates as a DNS sinkhole -- intercepting DNS queries from devices on the network and blocking requests to known advertising, tracking, and malware domains by returning null responses.

Unlike Pi-hole, which relies on a combination of shell scripts, PHP, dnsmasq, lighttpd, and SQLite, Mantis is a single Go binary that embeds all functionality: DNS server, DHCP server, recursive resolver, encrypted DNS (DoH/DoT), query logging, blocklist management, and a React-based admin dashboard. This eliminates dependency complexity and simplifies deployment to a single binary or Docker container.

Mantis targets network administrators, homelab enthusiasts, and privacy-conscious users who want a performant, modern ad blocker that scales from a Raspberry Pi to enterprise networks.

### 1.2 Target Audience

- **Homelab enthusiasts**: Self-hosting on Raspberry Pi, NUC, or home servers. Want single-binary deployment with minimal configuration.
- **Small/medium business IT admins**: Managing 50-500 devices across VLANs. Need reliable DNS+DHCP with centralized logging and statistics.
- **Enterprise network engineers**: Running 500+ device networks. Require high throughput, API-driven configuration, and monitoring integration.
- **Privacy-conscious users**: Want DNS encryption (DoH/DoT) and comprehensive tracker blocking without third-party cloud dependencies.

### 1.3 Key Differentiators

- **Single binary deployment**: All components (DNS, DHCP, web UI, API) compiled into one executable. No runtime dependencies, no interpreters, no web servers to configure.
- **Go-native performance**: Concurrent DNS resolution with goroutines. Sub-millisecond blocklist lookups via in-memory radix tree. Handles 100K+ queries/sec on commodity hardware.
- **Encrypted DNS built-in**: DoH and DoT as first-class features, not bolt-on additions. Both as a server (clients connect with encryption) and as an upstream client (queries forwarded with encryption).
- **Dual resolver mode**: Forwarding mode (relay to upstream like Cloudflare/Google) and recursive mode (resolve from root servers like unbound). User selects per deployment.
- **Modern admin UI**: React SPA with real-time query log, interactive charts, and dark mode. Embedded in the binary via Go's `embed` package.

### 1.4 Competitive Landscape

| Feature                  | Mantis        | Pi-hole          | AdGuard Home     | Blocky           |
|--------------------------|---------------|------------------|------------------|------------------|
| Language                 | Go            | PHP/Shell/C      | Go               | Go               |
| Single binary            | Yes           | No (multi-comp)  | Yes              | Yes              |
| DHCP server              | Yes           | Yes (dnsmasq)    | Yes              | No               |
| DNS-over-HTTPS (server)  | Yes           | No               | Yes              | No               |
| DNS-over-TLS (server)    | Yes           | No               | Yes              | No               |
| Recursive resolver       | Yes           | No (needs unbound)| No              | No               |
| Embedded web UI          | Yes (React)   | Yes (PHP)        | Yes (React)      | No               |
| Key-value store          | Pebble        | SQLite/FTL       | BoltDB           | None (in-memory) |
| Open source              | Yes           | Yes              | Yes              | Yes              |

## 2. Core Concepts

| Concept              | Definition                                                                                           |
|----------------------|------------------------------------------------------------------------------------------------------|
| DNS Sinkhole         | DNS server that returns NXDOMAIN or 0.0.0.0 for domains on a blocklist, preventing connection        |
| Blocklist            | Text file containing domains to block. Supports hosts-file format and domain-only format             |
| Allowlist            | Override list that permits specific domains even if they appear on a blocklist                        |
| Upstream             | External DNS server(s) that Mantis forwards queries to in forwarding mode                            |
| Recursive Resolution | Process of resolving a domain by querying root, TLD, and authoritative servers directly              |
| Client               | Any device on the network that sends DNS queries to Mantis                                           |
| Query Log            | Time-series record of every DNS query processed, including result, latency, and client info          |
| Gravity              | The compiled, deduplicated in-memory blocklist built from all configured blocklist sources            |
| DHCP Lease           | IP address assignment to a client device with an expiration time                                     |
| DoH                  | DNS-over-HTTPS: DNS queries tunneled through HTTPS on port 443                                       |
| DoT                  | DNS-over-TLS: DNS queries wrapped in TLS on port 853                                                 |

## 3. Functional Requirements

### 3.1 DNS Resolution

#### 3.1.1 Standard DNS Server

**User Story:** As a network admin, I want Mantis to respond to DNS queries on UDP/TCP port 53 so that all devices on my network can resolve domain names.

**Description:** Mantis listens on configurable interfaces and ports for DNS queries. For each incoming query, it checks the domain against Gravity (compiled blocklist). If blocked, it returns a null response. If allowed, it forwards the query to the configured upstream or resolves recursively.

**Acceptance Criteria:**
- [ ] Listens on UDP and TCP port 53 (configurable)
- [ ] Responds to A, AAAA, CNAME, MX, TXT, SRV, PTR, SOA, and NS query types
- [ ] Returns 0.0.0.0 (A) and :: (AAAA) for blocked domains
- [ ] Forwards non-blocked queries to upstream DNS servers
- [ ] Supports multiple upstream servers with failover
- [ ] Caches responses according to TTL from upstream
- [ ] Handles EDNS0 and supports responses up to 4096 bytes
- [ ] Responds within 5ms for cached/blocked queries (p99)

**Edge Cases:**
- Query for a CNAME that points to a blocked domain: block the final resolution
- Upstream server timeout: failover to next upstream within 2 seconds
- Malformed DNS packet: drop silently, increment error counter
- Query flood from single client: continue serving, log warning at threshold

**Constraints:**
- Maximum concurrent queries: 50,000
- DNS cache size: configurable, default 10,000 entries
- Cache TTL override: optional minimum/maximum TTL settings

#### 3.1.2 DNS-over-HTTPS (DoH) Server

**User Story:** As a privacy-conscious user, I want to query Mantis over HTTPS so that my DNS queries are encrypted in transit.

**Description:** Mantis serves a DoH endpoint on a configurable HTTPS port. Clients send DNS queries as HTTP POST (application/dns-message) or GET (dns parameter) requests per RFC 8484.

**Acceptance Criteria:**
- [ ] Serves DoH on configurable port (default 443)
- [ ] Supports POST with application/dns-message content type
- [ ] Supports GET with ?dns= base64url parameter
- [ ] Requires valid TLS certificate (auto-provision via Let's Encrypt or manual)
- [ ] Applies same blocking rules as standard DNS
- [ ] Returns proper HTTP status codes (200 for success, 400 for bad request)

**Edge Cases:**
- Invalid DNS wire format in HTTP body: return 400
- TLS certificate expiry: log error, attempt renewal, continue serving with expired cert
- Concurrent HTTPS connections exceeding limit: return 503

**Constraints:**
- TLS 1.2 minimum, TLS 1.3 preferred
- Maximum request body: 512 bytes

#### 3.1.3 DNS-over-TLS (DoT) Server

**User Story:** As a network admin, I want clients to connect to Mantis over TLS so that DNS traffic is encrypted on port 853.

**Description:** Mantis listens on port 853 for TLS-wrapped DNS queries per RFC 7858. Uses the same TLS certificate as DoH.

**Acceptance Criteria:**
- [ ] Listens on configurable port (default 853)
- [ ] Accepts TLS connections with DNS wire protocol inside
- [ ] Shares TLS certificate configuration with DoH
- [ ] Supports TCP keepalive for persistent connections
- [ ] Applies same blocking and logging as standard DNS

**Edge Cases:**
- Client sends plain DNS to DoT port: reject connection
- TLS handshake failure: log, increment counter, close connection

#### 3.1.4 Upstream Forwarding

**User Story:** As a user, I want to configure upstream DNS servers so that Mantis can resolve non-blocked queries.

**Description:** In forwarding mode, Mantis relays queries to configured upstream servers. Supports plain DNS, DoH, and DoT for upstream connections.

**Acceptance Criteria:**
- [ ] Supports multiple upstream servers with priority ordering
- [ ] Supports plain DNS (UDP/TCP), DoH, and DoT upstream protocols
- [ ] Implements parallel query to all upstreams, use fastest response
- [ ] Retries failed upstream queries with exponential backoff
- [ ] Health-checks upstream servers periodically (every 30s)
- [ ] Falls back to secondary upstream if primary is unhealthy

**Edge Cases:**
- All upstreams down: return SERVFAIL, log critical error
- Upstream returns SERVFAIL: try next upstream
- Upstream latency spike: timeout after configurable duration (default 5s)

#### 3.1.5 Recursive Resolution

**User Story:** As an advanced user, I want Mantis to resolve domains by querying root servers directly so that I don't depend on any third-party DNS provider.

**Description:** In recursive mode, Mantis performs iterative DNS resolution starting from root hints. It queries root servers, follows referrals through TLD and authoritative servers, and caches intermediate results.

**Acceptance Criteria:**
- [ ] Resolves queries starting from root hints (built-in, updatable)
- [ ] Follows NS referrals iteratively through the DNS hierarchy
- [ ] Validates DNSSEC signatures when available
- [ ] Caches intermediate NS records and glue records
- [ ] Applies blocklist filtering at every resolution step
- [ ] Configurable maximum recursion depth (default 30)

**Edge Cases:**
- Circular CNAME chain: detect and return SERVFAIL after max depth
- Lame delegation: detect, skip, try alternative NS
- DNSSEC validation failure: return SERVFAIL, log the chain break

**Constraints:**
- Root hints file embedded in binary, user-overridable
- Maximum recursion depth: 30 hops

### 3.2 Blocklist Management

#### 3.2.1 Blocklist Sources

**User Story:** As an admin, I want to add multiple blocklist URLs so that Mantis can aggregate blocking rules from various sources.

**Description:** Mantis downloads blocklists from configured URLs on a schedule, parses them into domain lists, deduplicates, and compiles them into Gravity (the in-memory lookup structure).

**Acceptance Criteria:**
- [ ] Supports hosts-file format (127.0.0.1 domain.com or 0.0.0.0 domain.com)
- [ ] Supports domain-only format (one domain per line)
- [ ] Supports comment lines (# prefix)
- [ ] Downloads lists via HTTP/HTTPS
- [ ] Configurable update interval (default: daily)
- [ ] Ships with 5+ default blocklist URLs (StevenBlack, EasyList, etc.)
- [ ] Deduplicates across all lists
- [ ] Reports per-list domain count in admin UI

**Edge Cases:**
- Blocklist URL returns 404/500: keep previous version, log warning
- Blocklist download timeout: retry 3 times with backoff
- Corrupt/unparseable blocklist: skip invalid lines, report parse errors
- Blocklist with 1M+ domains: handle without excessive memory

**Constraints:**
- Maximum blocklist size per source: 50MB
- Maximum total unique domains in Gravity: 5,000,000

#### 3.2.2 Allowlist

**User Story:** As an admin, I want to allowlist specific domains so that they are never blocked even if they appear on a blocklist.

**Description:** Allowlist entries override blocklist matches. Supports exact domain matches and wildcard patterns.

**Acceptance Criteria:**
- [ ] Exact domain matching (e.g., `ads.example.com`)
- [ ] Wildcard matching (e.g., `*.example.com` matches all subdomains)
- [ ] Regex matching for advanced patterns
- [ ] Allowlist checked before blocklist on every query
- [ ] Manageable via admin UI and API
- [ ] Changes take effect immediately without restart

**Edge Cases:**
- Allowlisted domain is a CNAME to a blocked domain: allow the entire chain
- Regex pattern causes catastrophic backtracking: timeout regex evaluation at 10ms

#### 3.2.3 Custom Block Rules

**User Story:** As an admin, I want to add my own custom block rules beyond the subscribed lists.

**Description:** Users can add individual domains or patterns to a local blocklist that persists independently of subscribed lists.

**Acceptance Criteria:**
- [ ] Add/remove individual domains via UI and API
- [ ] Support wildcard blocking (e.g., `*.ads.example.com`)
- [ ] Custom rules survive Gravity rebuild
- [ ] Import/export custom rules as text file
- [ ] Changes take effect immediately

### 3.3 DHCP Server

#### 3.3.1 IPv4 DHCP

**User Story:** As a network admin, I want Mantis to serve as my DHCP server so that all devices automatically use Mantis as their DNS server.

**Description:** Mantis includes a DHCPv4 server that assigns IP addresses from a configurable pool. It automatically sets itself as the DNS server in DHCP offers.

**Acceptance Criteria:**
- [ ] Assigns IPv4 addresses from configurable range
- [ ] Configurable lease duration (default 24 hours)
- [ ] Sets gateway, subnet mask, DNS server in DHCP options
- [ ] Supports static/reserved leases by MAC address
- [ ] Displays active leases in admin UI
- [ ] Sends DHCP NAK for out-of-range requests
- [ ] Handles DHCP DISCOVER, OFFER, REQUEST, ACK, NAK, RELEASE, INFORM

**Edge Cases:**
- Address pool exhaustion: return NAK, log critical warning
- Duplicate MAC with different hostname: update hostname, keep reservation
- Client requests IP outside pool: offer from pool instead

**Constraints:**
- Maximum lease pool size: /16 subnet (65,534 addresses)
- Lease persistence: stored in Pebble, survives restart

#### 3.3.2 IPv6 DHCP (DHCPv6) -- Deferred to v1.1

> **Note:** DHCPv6 is deferred to v1.1. v1.0 ships with DHCPv4 only.

### 3.4 Query Logging and Statistics

#### 3.4.1 Query Log

**User Story:** As an admin, I want to see every DNS query processed by Mantis so that I can troubleshoot network issues and monitor activity.

**Description:** Every DNS query is logged with timestamp, client IP, queried domain, query type, result (allowed/blocked/cached), upstream used, and response latency.

**Acceptance Criteria:**
- [ ] Logs all DNS queries with full metadata
- [ ] Stores logs in Pebble with configurable retention (default 30 days)
- [ ] Query log searchable by domain, client, result type, and time range
- [ ] Real-time query log stream via WebSocket in admin UI
- [ ] Supports log export as CSV
- [ ] Configurable log level: all queries, blocked only, or disabled
- [ ] Privacy mode: option to anonymize client IPs in logs

**Edge Cases:**
- High query volume (10K+ queries/sec): batch writes, accept eventual consistency
- Storage full: delete oldest entries, log warning
- Corrupt log entry: skip on read, report in diagnostics

**Constraints:**
- Maximum log retention: 365 days
- Log entry size: approximately 200 bytes per query
- Storage estimate: 1.7 GB per million queries

#### 3.4.2 Statistics Dashboard

**User Story:** As an admin, I want to see summary statistics about DNS activity so that I can understand blocking effectiveness and network patterns.

**Description:** Admin UI displays aggregated statistics computed from the query log.

**Acceptance Criteria:**
- [ ] Total queries in last 24h, 7d, 30d
- [ ] Percentage of queries blocked
- [ ] Top 10 queried domains (allowed and blocked)
- [ ] Top 10 clients by query volume
- [ ] Queries over time chart (hourly resolution for 24h, daily for 30d)
- [ ] Upstream response time distribution
- [ ] Cache hit ratio
- [ ] Statistics computed from Pebble aggregates, not full table scans

### 3.5 Admin Interface

#### 3.5.1 Web Dashboard

**User Story:** As an admin, I want a web-based dashboard to monitor and configure Mantis.

**Description:** React SPA served by Mantis's embedded HTTP server. Provides real-time monitoring, configuration management, and query log access.

**Key Screens:**
1. **Dashboard**: Summary cards (total queries, blocked %, cache hit rate, active clients), query rate chart, top blocked domains
2. **Query Log**: Searchable, filterable real-time table with domain, client, type, result, latency columns. WebSocket live stream toggle
3. **Blocklists**: List of subscribed blocklists with domain counts, last update time. Add/remove/toggle lists. Manual Gravity rebuild trigger
4. **Allow/Block Lists**: Manage custom allow and block rules. Add/remove domains with wildcard support
5. **DHCP**: Active leases table, static lease management, pool configuration
6. **Network**: Client list with hostnames, IPs, MACs, query counts. Per-client blocking toggle
7. **Settings**: General config (upstream DNS, listening interfaces, ports), DNS encryption settings (DoH/DoT certificates), logging preferences, system info
8. **Long-term Statistics**: Historical charts and trends beyond 24h

**Acceptance Criteria:**
- [ ] Responsive design (desktop and tablet)
- [ ] Dark mode and light mode
- [ ] Session-based authentication (username/password)
- [ ] All configuration changes apply without restart
- [ ] Loading states and error handling on all API calls
- [ ] Keyboard navigable for accessibility

#### 3.5.2 API

**User Story:** As a power user, I want a REST API to automate Mantis configuration and monitoring.

**Description:** All admin UI functionality exposed through a versioned REST API.

**Acceptance Criteria:**
- [ ] Versioned under /api/v1
- [ ] JSON request/response format
- [ ] API key authentication (generated from admin UI)
- [ ] Rate limited to prevent abuse (default 60 req/min)
- [ ] OpenAPI/Swagger specification auto-generated

### 3.6 Configuration

#### 3.6.1 Configuration Sources

**User Story:** As a user, I want multiple ways to configure Mantis depending on my deployment scenario.

**Description:** Mantis supports configuration via TOML file, environment variables, and CLI flags. Priority: CLI flags > env vars > config file > defaults.

**Acceptance Criteria:**
- [ ] TOML configuration file (default: /etc/mantis/config.toml)
- [ ] Environment variables with MANTIS_ prefix
- [ ] CLI flags for all critical settings
- [ ] Config file path configurable via --config flag
- [ ] Validation on startup with clear error messages
- [ ] Example config file shipped with distribution

## 4. Architecture Overview

### 4.1 System Components

| Component         | Responsibility                                                      |
|-------------------|---------------------------------------------------------------------|
| DNS Engine        | Listens for queries, applies blocking, forwards/resolves, caches    |
| DoH/DoT Server    | TLS termination, HTTP/2 handling, DNS wire format over HTTPS/TLS    |
| Recursive Resolver| Iterative resolution from root hints, DNSSEC validation             |
| DHCP Server       | DHCPv4 and DHCPv6 address assignment and lease management           |
| Gravity Engine    | Blocklist download, parse, deduplicate, compile to in-memory lookup |
| Query Logger      | Async write of query metadata to Pebble                           |
| Statistics Engine  | Aggregation of query log into time-bucketed statistics              |
| API Server        | REST endpoints for admin UI and automation                          |
| Web Server        | Serves embedded React SPA, proxies API requests                     |
| Config Manager    | Loads, validates, and hot-reloads configuration                     |
| Storage Layer     | Pebble abstraction for logs, leases, settings, and blocklists     |

### 4.2 Component Interactions

```
                    +-----------+
                    |  Clients  |
                    +-----+-----+
                          |
           +--------------+--------------+
           |              |              |
      UDP/TCP:53     HTTPS:443      TLS:853
           |              |              |
    +------+------+ +-----+-----+ +-----+-----+
    | DNS Engine  | | DoH Server| | DoT Server|
    +------+------+ +-----+-----+ +-----+-----+
           |              |              |
           +-------+------+------+-------+
                   |                     
           +-------v--------+           
           | Gravity Engine |           
           | (blocklist     |           
           |  lookup)       |           
           +-------+--------+           
                   |                     
          blocked? |                     
        +----+-----+-----+              
        |                 |              
   +----v----+    +-------v--------+    
   | Return  |    | Forward/Resolve|    
   | 0.0.0.0 |    +-------+--------+    
   +---------+            |              
                   +------v------+       
                   | DNS Cache   |       
                   +------+------+       
                          |              
                   +------v------+       
                   | Query Logger|       
                   +------+------+       
                          |              
                   +------v------+       
                   |  Pebble   |       
                   +-------------+       
```

Admin path:
```
Browser → Web Server (React SPA) → API Server → [Config Manager | Storage Layer | Gravity Engine]
```

DHCP path:
```
Client DHCP Request → DHCP Server → Pebble (leases)
```

### 4.3 External Integrations

| Integration        | Purpose                                      | Fallback                               |
|--------------------|----------------------------------------------|----------------------------------------|
| Upstream DNS       | Forward queries to external resolvers         | Switch to recursive mode or next upstream |
| Blocklist URLs     | Download domain blocklists                    | Use cached version from last download  |
| Let's Encrypt      | Auto-provision TLS certificates for DoH/DoT   | Manual certificate configuration       |
| NTP                | Accurate timestamps for DNSSEC validation     | System clock (warn if skewed)          |
| Root Hints         | Starting point for recursive resolution       | Embedded root hints file               |

## 5. Data Model

### 5.1 Core Entities

#### BlocklistSource

| Field       | Type     | Required | Description                        | Constraints                   |
|-------------|----------|----------|------------------------------------|-------------------------------|
| id          | string   | Yes      | Unique identifier                  | UUID v4, auto-generated       |
| name        | string   | Yes      | Display name                       | 1-256 characters              |
| url         | string   | Yes      | Download URL                       | Valid HTTP/HTTPS URL          |
| enabled     | bool     | Yes      | Whether this source is active      | Default: true                 |
| format      | string   | Yes      | hosts, domains, or adblock         | Enum                          |
| domainCount | int      | No       | Number of domains in this list     | Updated on Gravity rebuild    |
| lastUpdated | datetime | No       | Last successful download timestamp | UTC                           |
| lastStatus  | string   | No       | Last download result               | success, error, pending       |

#### CustomRule

| Field   | Type     | Required | Description                       | Constraints                   |
|---------|----------|----------|-----------------------------------|-------------------------------|
| id      | string   | Yes      | Unique identifier                 | UUID v4, auto-generated       |
| domain  | string   | Yes      | Domain or pattern                 | Valid domain or wildcard       |
| type    | string   | Yes      | block or allow                    | Enum                          |
| comment | string   | No       | Admin note                        | 0-512 characters              |
| created | datetime | Yes      | Creation timestamp                | UTC                           |

#### QueryLogEntry

| Field      | Type     | Required | Description                       | Constraints                   |
|------------|----------|----------|-----------------------------------|-------------------------------|
| id         | uint64   | Yes      | Auto-increment ID                 | Time-ordered                  |
| timestamp  | datetime | Yes      | Query timestamp                   | UTC, nanosecond precision     |
| clientIp   | string   | Yes      | Client IP address                 | IPv4 or IPv6                  |
| domain     | string   | Yes      | Queried domain                    | Lowercased, no trailing dot   |
| queryType  | uint16   | Yes      | DNS query type (A, AAAA, etc.)    | DNS QTYPE value               |
| result     | string   | Yes      | allowed, blocked, cached, error   | Enum                          |
| upstream   | string   | No       | Upstream server used              | IP:port or resolver           |
| latencyUs  | int64    | Yes      | Response latency in microseconds  | >= 0                          |
| answer     | string   | No       | Resolved IP or CNAME              | Truncated at 256 chars        |

#### DhcpLease

| Field     | Type     | Required | Description                       | Constraints                   |
|-----------|----------|----------|-----------------------------------|-------------------------------|
| mac       | string   | Yes      | Client MAC address                | Canonical format AA:BB:CC...  |
| ip        | string   | Yes      | Assigned IP address               | IPv4 or IPv6                  |
| hostname  | string   | No       | Client hostname                   | From DHCP option 12           |
| leaseEnd  | datetime | Yes      | Lease expiration                  | UTC                           |
| isStatic  | bool     | Yes      | Whether this is a static lease    | Default: false                |

#### Settings

| Field | Type   | Required | Description         | Constraints         |
|-------|--------|----------|---------------------|---------------------|
| key   | string | Yes      | Setting key         | Dot-notation path   |
| value | string | Yes      | Serialized value    | JSON-encoded        |

### 5.2 Relationships

- BlocklistSource --> contributes to --> Gravity (compiled in-memory)
- CustomRule --> merges into --> Gravity (allow rules checked first)
- QueryLogEntry --> references --> Client (by clientIp)
- DhcpLease --> identifies --> Client (by mac)
- Settings --> configures --> all components

### 5.3 Data Lifecycle

- **QueryLogEntry**: Created on every query. Deleted after retention period (configurable, default 30 days). Bulk deletion via daily cleanup job.
- **DhcpLease**: Created on DHCP ACK. Updated on renewal. Deleted on expiry or release. Static leases persist until manually removed.
- **BlocklistSource**: Created by admin. Updated on each Gravity rebuild. Deleted by admin.
- **CustomRule**: Created/deleted by admin. No expiry.
- **Settings**: Created on first configuration. Updated via API/UI. Never deleted (reset to defaults).

## 6. API Surface

### 6.1 API Style

REST with JSON. Versioned under /api/v1. Chosen for simplicity, broad client compatibility, and ease of scripting with curl.

### 6.2 Endpoint Overview

| Method | Path                                  | Description                          | Auth     |
|--------|---------------------------------------|--------------------------------------|----------|
| POST   | /api/v1/auth/login                    | Authenticate, receive session token  | Public   |
| POST   | /api/v1/auth/logout                   | Invalidate session                   | Required |
| GET    | /api/v1/stats/summary                 | Dashboard summary statistics         | Required |
| GET    | /api/v1/stats/overtime                | Queries over time chart data         | Required |
| GET    | /api/v1/stats/top-domains             | Top queried/blocked domains          | Required |
| GET    | /api/v1/stats/top-clients             | Top clients by query volume          | Required |
| GET    | /api/v1/queries                       | Query log with pagination/filters    | Required |
| GET    | /api/v1/queries/stream                | WebSocket live query stream          | Required |
| GET    | /api/v1/blocklists                    | List all blocklist sources           | Required |
| POST   | /api/v1/blocklists                    | Add blocklist source                 | Required |
| PUT    | /api/v1/blocklists/:id                | Update blocklist source              | Required |
| DELETE | /api/v1/blocklists/:id                | Remove blocklist source              | Required |
| POST   | /api/v1/gravity/rebuild               | Trigger Gravity rebuild              | Required |
| GET    | /api/v1/gravity/status                | Gravity build status and stats       | Required |
| GET    | /api/v1/rules                         | List custom allow/block rules        | Required |
| POST   | /api/v1/rules                         | Add custom rule                      | Required |
| DELETE | /api/v1/rules/:id                     | Remove custom rule                   | Required |
| GET    | /api/v1/dhcp/leases                   | List DHCP leases                     | Required |
| POST   | /api/v1/dhcp/leases/static            | Add static lease                     | Required |
| DELETE | /api/v1/dhcp/leases/static/:mac       | Remove static lease                  | Required |
| GET    | /api/v1/dhcp/config                   | Get DHCP configuration               | Required |
| PUT    | /api/v1/dhcp/config                   | Update DHCP configuration            | Required |
| GET    | /api/v1/clients                       | List known clients                   | Required |
| PUT    | /api/v1/clients/:ip/blocking          | Toggle blocking for specific client  | Required |
| GET    | /api/v1/settings                      | Get all settings                     | Required |
| PUT    | /api/v1/settings                      | Update settings                      | Required |
| GET    | /api/v1/system/info                   | System version, uptime, resource use | Required |
| POST   | /api/v1/system/restart-dns            | Restart DNS engine                   | Required |
| GET    | /api/v1/dns/test                      | Test DNS resolution for a domain     | Required |

### 6.3 Authentication and Authorization

- **Method**: Session-based with HTTP-only cookie
- **API keys**: Long-lived tokens for automation, generated from admin UI
- **Single admin user**: v1.0 supports one admin account (multi-user in future)
- **Password**: bcrypt hashed, minimum 8 characters
- **Session duration**: 24 hours, sliding window

### 6.4 Rate Limiting

- API: 60 requests/minute per session (configurable)
- Login: 5 attempts/minute per IP
- Rate limit headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset

### 6.5 Error Format

```json
{
  "error": {
    "code": "BLOCKLIST_DOWNLOAD_FAILED",
    "message": "Failed to download blocklist from https://example.com/list.txt: connection timeout"
  }
}
```

Standard HTTP status codes: 200, 201, 204, 400, 401, 403, 404, 429, 500.

## 7. User Interface

### 7.1 Interface Type

React SPA (TypeScript) embedded in Go binary via `embed` package. Served on the same port as the API (default :8080).

### 7.2 Key Screens

1. **Dashboard** (/)
   - Purpose: At-a-glance network and blocking overview
   - Elements: Summary cards, 24h query chart, top blocked/allowed domains, active client count
   - Actions: Quick toggle blocking on/off, navigate to details
   - Navigation: Sidebar link, default landing page

2. **Query Log** (/queries)
   - Purpose: Real-time and historical DNS query inspection
   - Elements: Filterable table, search bar, live stream toggle, pagination
   - Actions: Search, filter by type/result/client, allow/block domain inline, export
   - Navigation: Sidebar link, click-through from dashboard

3. **Blocklist Management** (/blocklists)
   - Purpose: Manage subscribed blocklist sources
   - Elements: Source list with stats, add form, enable/disable toggles
   - Actions: Add/remove sources, toggle enable, trigger Gravity rebuild, view per-list stats
   - Navigation: Sidebar link

4. **Custom Rules** (/rules)
   - Purpose: Manage per-domain allow and block overrides
   - Elements: Rules table with search, add form, import/export buttons
   - Actions: Add/remove rules, toggle allow/block type, bulk import
   - Navigation: Sidebar link

5. **DHCP** (/dhcp)
   - Purpose: DHCP server management
   - Elements: Active leases table, static leases table, pool config form
   - Actions: Add/remove static leases, modify pool range, view lease history
   - Navigation: Sidebar link

6. **Network** (/network)
   - Purpose: Client device overview
   - Elements: Client table with hostname/IP/MAC/query count, per-client blocking toggle
   - Actions: Rename clients, toggle blocking per client, view client query history
   - Navigation: Sidebar link, click-through from dashboard

7. **Settings** (/settings)
   - Purpose: System configuration
   - Elements: Tabbed form (DNS, DHCP, Encryption, Logging, System)
   - Actions: Modify all settings, test DNS resolution, manage API keys
   - Navigation: Sidebar link

8. **Long-term Statistics** (/stats)
   - Purpose: Historical trends and patterns
   - Elements: Date range selector, charts (queries/day, block rate, top domains over time)
   - Actions: Select time range, export data
   - Navigation: Sidebar link

### 7.3 Responsive Requirements

- Desktop-first design (primary use case: admin on laptop/desktop)
- Responsive down to 768px width (tablet)
- Mobile (< 768px): simplified layout, collapsible sidebar
- WCAG 2.1 AA compliance
- Keyboard navigation support
- Dark mode as default, light mode available

## 8. Security Model

### 8.1 Authentication

- Single admin account configured on first run (setup wizard)
- bcrypt password hashing with cost factor 12
- Session tokens: cryptographically random, 256-bit
- HTTP-only, Secure, SameSite=Strict cookies
- API keys: 64-character random tokens, stored hashed

### 8.2 Authorization

- v1.0: Single role (admin) -- full access
- API key: same permissions as admin session
- Future: read-only role for monitoring dashboards

### 8.3 Data Protection

- TLS for DoH/DoT (user-provided or Let's Encrypt certificate)
- Admin UI served over HTTPS when TLS is configured
- Pebble encrypted at rest with AES-256 (optional, configurable)
- No PII transmitted to external services
- Query log anonymization option (mask last octet of client IP)

### 8.4 Input Validation

- All API inputs validated and sanitized
- Domain names validated against RFC 1035
- IP addresses validated for format and range
- Config file parsed with strict schema validation
- No shell command execution from user input

## 9. Deployment Model

### 9.1 Target Environments

- Raspberry Pi (ARM64) -- primary target for home users
- Linux x86_64 servers -- primary target for business/enterprise
- macOS (development and small office use)
- Docker on any platform

### 9.2 Distribution Method

1. **Single binary**: Download from GitHub releases. Available for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
2. **Docker image**: Multi-arch image on Docker Hub / GitHub Container Registry
3. **Docker Compose**: Example compose file with recommended settings
4. **Install script**: One-line curl installer for Linux (creates systemd service)

### 9.3 Configuration

- Primary: /etc/mantis/config.toml (Linux) or ./config.toml (Docker/dev)
- Override: MANTIS_* environment variables
- Override: CLI flags (--dns-port, --api-port, --config, etc.)
- Runtime changes: Via admin UI/API, persisted to Pebble

### 9.4 System Requirements

| Environment   | CPU       | RAM       | Storage   |
|---------------|-----------|-----------|-----------|
| Minimum       | 1 core    | 128 MB    | 256 MB    |
| Recommended   | 2 cores   | 512 MB    | 1 GB      |
| Enterprise    | 4+ cores  | 2+ GB     | 10+ GB    |

- OS: Linux (kernel 4.4+), macOS 12+
- Ports: 53 (DNS), 67-68 (DHCP), 443 (DoH), 853 (DoT), 8080 (Admin)
- Privileges: Root or CAP_NET_BIND_SERVICE for ports < 1024

## 10. Performance Requirements

### 10.1 Response Time Targets

| Operation              | p50     | p95     | p99      |
|------------------------|---------|---------|----------|
| Blocked query          | < 0.5ms | < 1ms   | < 5ms    |
| Cached query           | < 0.5ms | < 1ms   | < 5ms    |
| Forwarded query        | < 20ms  | < 50ms  | < 100ms  |
| Recursive query        | < 50ms  | < 200ms | < 500ms  |
| API endpoint           | < 10ms  | < 50ms  | < 200ms  |
| Admin UI page load     | < 500ms | < 1s    | < 2s     |

### 10.2 Throughput Targets

- DNS queries: 100,000 queries/sec on 4-core machine
- API requests: 1,000 requests/sec
- WebSocket connections: 100 concurrent live query streams
- DHCP operations: 1,000 leases/sec

### 10.3 Resource Limits

- Idle memory: < 50 MB (without large blocklists)
- Memory with 2M blocked domains: < 256 MB
- Memory with 5M blocked domains: < 512 MB
- CPU idle: < 1% of single core
- Pebble disk growth: approximately 2 GB per million query log entries

## 11. Constraints and Non-Goals

### 11.1 Technical Constraints

- Go 1.22+ required for build
- CGo-free build (pure Go, cross-compilation friendly)
- Must run on Raspberry Pi Zero 2 W (ARM64, 512MB RAM)
- Must handle graceful shutdown (drain in-flight queries)
- No external runtime dependencies (no Python, no Node.js, no PHP)

### 11.2 Non-Goals

- **VPN/proxy functionality**: Mantis is DNS-only, not a network proxy
- **Content filtering**: No HTTP-level inspection, URL filtering, or SSL interception
- **Multi-node clustering**: v1.0 is single-instance only
- **Email notifications**: No alerting system in v1.0
- **Plugin/extension system**: No third-party plugin architecture
- **Mobile native apps**: Web UI only, no iOS/Android app
- **GUI installer**: CLI and Docker only, no graphical installer
- **Automatic network detection**: User must configure interfaces manually
- **Full recursive DNSSEC enforcement**: Validate when possible, don't hard-fail on missing signatures
- **Commercial support tier**: Open source only

### 11.3 Assumptions

- Network allows Mantis to bind to required ports
- Upstream DNS servers are reachable from the deployment host
- System clock is approximately accurate (within 5 minutes for DNSSEC)
- Users have basic networking knowledge (can configure static IP, understand DNS)

### 11.4 Open Questions

All open questions have been resolved:

- **Pebble** selected as KV store (CockroachDB's LSM engine, write-optimized, stable)
- **Built-in ACME client** for Let's Encrypt auto-provisioning in v1.0
- **DHCPv6 deferred to v1.1**, v1.0 ships with DHCPv4 only

## 12. Future Considerations

- **v1.1**: DHCPv6 (stateful IPv6 address assignment)
- **v1.1**: Multi-user support with RBAC (admin, viewer roles)
- **v1.1**: Email/webhook alerting for critical events (upstream down, disk full)
- **v1.2**: Conditional forwarding (route specific domains to specific upstreams)
- **v1.2**: Prometheus metrics endpoint for monitoring integration
- **v2.0**: Multi-node synchronization (primary-replica blocklist and config sync)
- **v2.0**: Plugin system for custom DNS processing logic
- **v2.0**: Mobile-optimized PWA or native companion app
