---
title: drift init
description: Initialize a new drift project.
---

# drift init

Create a new drift project with a config file and migrations directory.

## Usage

```bash
drift init [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--driver DRIVER` | Database driver: `postgres`, `mysql`, `sqlite` |

## Examples

```bash
# Initialize with default settings
drift init

# Initialize for PostgreSQL
drift init --driver postgres

# Initialize for SQLite
drift init --driver sqlite
```

## What It Creates

```
./
├── drift.yaml       # Configuration file with sensible defaults
└── migrations/      # Directory for migration files
```

The generated `drift.yaml` includes a URL template appropriate for the selected driver.
