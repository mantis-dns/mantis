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
