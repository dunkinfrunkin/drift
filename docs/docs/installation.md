---
sidebar_position: 2
title: Installation
description: Install drift via Homebrew, curl, Go, or Docker.
---

# Installation

Drift ships as a single binary with zero runtime dependencies. Choose your preferred method:

## Homebrew (macOS / Linux)

```bash
brew install dunkinfrunkin/tap/drift
```

## Shell Script

```bash
curl -fsSL https://raw.githubusercontent.com/dunkinfrunkin/drift/main/install.sh | sh
```

This auto-detects your OS and architecture, downloads the correct binary, and installs to `/usr/local/bin`.

## Go Install

Requires Go 1.23+:

```bash
go install github.com/frankchan/drift/cmd/drift@latest
```

## Docker

```bash
docker pull ghcr.io/dunkinfrunkin/drift:latest
```

Usage with Docker:

```bash
docker run --rm \
  -v $(pwd)/migrations:/migrations \
  ghcr.io/dunkinfrunkin/drift:latest \
  migrate --url postgres://host.docker.internal:5432/mydb
```

## GitHub Releases

Download pre-built binaries directly from [GitHub Releases](https://github.com/dunkinfrunkin/drift/releases):

| Platform | Architecture | File |
|----------|-------------|------|
| macOS | Apple Silicon (M1/M2/M3) | `drift_darwin_arm64.tar.gz` |
| macOS | Intel | `drift_darwin_amd64.tar.gz` |
| Linux | ARM64 | `drift_linux_arm64.tar.gz` |
| Linux | x86_64 | `drift_linux_amd64.tar.gz` |
| Windows | x86_64 | `drift_windows_amd64.zip` |
| Windows | ARM64 | `drift_windows_arm64.zip` |

## Verify Installation

```bash
drift version
```

```
drift version 0.1.1 (commit: abc8bf9, built: 2025-01-15T10:00:00Z)
```
