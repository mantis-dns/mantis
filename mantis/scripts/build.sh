#!/bin/bash
set -euo pipefail

VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
LDFLAGS="-s -w -X main.version=${VERSION}"
OUTPUT_DIR="dist"

echo "Building Mantis ${VERSION}..."

# Build frontend.
echo "Building frontend..."
cd web && npm ci && npm run build && cd ..
rm -rf internal/web/dist
cp -r web/dist internal/web/dist

# Cross-compile.
mkdir -p "$OUTPUT_DIR"
platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64")

for platform in "${platforms[@]}"; do
  IFS='/' read -r os arch <<< "$platform"
  output="${OUTPUT_DIR}/mantis-${os}-${arch}"
  echo "  Building ${os}/${arch}..."
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -ldflags="$LDFLAGS" -o "$output" ./cmd/mantis
done

# Checksums.
cd "$OUTPUT_DIR"
sha256sum mantis-* > checksums.txt
cd ..

echo ""
echo "Build complete:"
ls -lh "$OUTPUT_DIR"/mantis-*
echo ""
cat "$OUTPUT_DIR/checksums.txt"
