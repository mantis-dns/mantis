#!/bin/bash
set -euo pipefail

MANTIS_VERSION="${1:-latest}"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/mantis"
DATA_DIR="/var/lib/mantis"

echo "Installing Mantis ${MANTIS_VERSION}..."

# Detect architecture.
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
BINARY="mantis-${OS}-${ARCH}"

echo "Detected: ${OS}/${ARCH}"

# Create mantis user.
if ! id -u mantis &>/dev/null; then
  useradd --system --no-create-home --shell /usr/sbin/nologin mantis
  echo "Created mantis user"
fi

# Create directories.
mkdir -p "$CONFIG_DIR" "$DATA_DIR"
chown mantis:mantis "$DATA_DIR"

# Copy default config if not exists.
if [ ! -f "$CONFIG_DIR/config.toml" ]; then
  cat > "$CONFIG_DIR/config.toml" << 'TOML'
[dns]
listen_address = "0.0.0.0:53"
upstreams = ["1.1.1.1", "8.8.8.8", "9.9.9.9"]
resolver_mode = "forward"
cache_size = 10000
blocking_mode = "null"

[api]
listen_address = "0.0.0.0:8080"

[storage]
data_dir = "/var/lib/mantis"

[logging]
level = "info"
query_log = "all"
retention_days = 30
TOML
  echo "Created default config at $CONFIG_DIR/config.toml"
fi

# Install systemd service.
if [ -d /etc/systemd/system ]; then
  cat > /etc/systemd/system/mantis.service << 'SERVICE'
[Unit]
Description=Mantis DNS Sinkhole
After=network.target

[Service]
Type=simple
User=mantis
Group=mantis
ExecStart=/usr/local/bin/mantis --config /etc/mantis/config.toml --data-dir /var/lib/mantis
Restart=always
RestartSec=5
LimitNOFILE=65535
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
SERVICE
  systemctl daemon-reload
  echo "Installed systemd service"
fi

echo ""
echo "Mantis installed successfully!"
echo ""
echo "Next steps:"
echo "  1. Edit config: $CONFIG_DIR/config.toml"
echo "  2. Start: systemctl start mantis"
echo "  3. Enable on boot: systemctl enable mantis"
echo "  4. Open admin: http://$(hostname -I | awk '{print $1}'):8080"
