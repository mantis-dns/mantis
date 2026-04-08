# Mantis

Network-level ad and tracker blocker. A modern DNS sinkhole written in Go.

Single binary. DNS server (UDP/TCP/DoH/DoT), recursive resolver, DHCPv4, query logging, blocklist management, and embedded React admin dashboard.

## Quick Start

```bash
make build
./dist/mantis --config configs/mantis.example.toml --data-dir ./data
```

## Build

```bash
# Development
make dev

# Production binary
make build

# All platforms
make build-all

# Run tests
make test
```

## Configuration

Copy and edit the example config:

```bash
cp configs/mantis.example.toml /etc/mantis/config.toml
```

All settings can be overridden with `MANTIS_` environment variables.

## License

MIT
