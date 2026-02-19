---
sidebar_position: 9
title: Building from Source
description: Build drift from source with Go and Node.js.
---

# Building from Source

## Prerequisites

- **Go 1.23+**
- **Node.js 20+** (for building the web UI)
- **Make** (optional, for convenience)

## Build

```bash
# Clone the repository
git clone https://github.com/dunkinfrunkin/drift.git
cd drift

# Build everything (UI + Go binary)
make build

# Or step by step:
cd ui && npm ci && npm run build && cd ..
go build -o bin/drift ./cmd/drift
```

## Install Locally

```bash
make install
# Or:
go install ./cmd/drift
```

## Development

```bash
# Run tests
make test

# Run tests with race detection
go test ./... -count=1 -race

# Lint
make lint

# Format code
make fmt

# Build UI in dev mode (with hot reload)
cd ui && npm run dev
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build UI and Go binary |
| `make build-go` | Build Go binary only |
| `make install` | Install to `$GOPATH/bin` |
| `make ui` | Build the web UI |
| `make test` | Run all tests |
| `make test-integration` | Run integration tests |
| `make lint` | Run Go vet |
| `make fmt` | Format Go code |
| `make clean` | Remove build artifacts |

## Project Structure

```
drift/
├── cmd/drift/          # CLI entry point
├── internal/
│   ├── cli/            # Cobra command handlers
│   ├── config/         # YAML config loading
│   ├── database/       # Database abstraction + drivers
│   │   ├── postgres/
│   │   ├── mysql/
│   │   └── sqlite/
│   ├── diff/           # Schema diffing
│   ├── engine/         # Core migration engine
│   ├── history/        # History table CRUD
│   ├── lint/           # SQL linting
│   ├── migration/      # File parsing
│   └── ui/             # Embedded web server
├── ui/                 # React + Vite frontend
├── docs/               # Documentation (Docusaurus)
└── migrations/         # Example migrations
```

## Release

Releases are automated with [GoReleaser](https://goreleaser.com/). Pushing a tag triggers the CI pipeline:

```bash
git tag -a v0.2.0 -m "v0.2.0"
git push origin v0.2.0
```

This builds binaries for 6 platforms, pushes Docker images to GHCR, and updates the Homebrew formula automatically.
