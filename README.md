# Mantis

Network-level ad and tracker blocker. A modern DNS sinkhole written in Go.

Single binary. DNS server (UDP/TCP/DoH/DoT), recursive resolver, DHCPv4, query logging, blocklist management, and embedded React admin dashboard.

## Features

- DNS sinkhole with gravity blocklist engine (radix tree, atomic swap)
- Forwarding and recursive resolver modes
- DNS-over-HTTPS (DoH) and DNS-over-TLS (DoT)
- DHCPv4 server with static lease management
- Real-time query log and statistics dashboard
- Blocklist management with scheduled rebuilds
- Custom allowlist/denylist rules
- Embedded React SPA admin panel
- Pebble embedded KV store (no external database)

## Quick Start

```bash
make build
./dist/mantis --config configs/mantis.example.toml --data-dir ./data
```

## Docker

```bash
docker compose -f docker/docker-compose.yml up -d
```

Data persists in the `mantis-data` volume. Config via `MANTIS_*` environment variables.

## Build

```bash
# Development
make dev

# Production binary
make build

# All platforms (linux/darwin, amd64/arm64)
make build-all

# Run tests
make test
```

## Configuration

Copy and edit the example config:

```bash
cp configs/mantis.example.toml /etc/mantis/config.toml
```

All settings can be overridden with `MANTIS_` environment variables (e.g. `MANTIS_DNS_LISTEN_ADDRESS=0.0.0.0:53`).

## License

MIT
