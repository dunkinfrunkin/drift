---
title: drift squash
description: Squash migrations into a single file.
---

# drift squash

Consolidate multiple versioned migration files into a single SQL file. Useful for cleaning up a long migration history.

## Usage

```bash
drift squash [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--up-to VERSION` | Squash up to this version (inclusive) |
| `--output FILE` | Write output to file (default: stdout) |

## Examples

```bash
# Squash all migrations up to V010
drift squash --up-to V010 --output migrations/V000__squashed.sql

# Preview squashed output
drift squash --up-to V005
```

## Workflow

1. Run `drift squash` to generate a consolidated file
2. Replace the individual migration files with the squashed file
3. Run `drift baseline` so existing databases skip the squashed version
4. New environments get a clean single-file setup
