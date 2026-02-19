---
sidebar_position: 3
title: Configuration
description: Configure drift with YAML and environment variables.
---

# Configuration

Drift uses a `drift.yaml` file for project configuration. All settings can also be overridden via CLI flags.

## Configuration File

Create a `drift.yaml` in your project root (or use `drift init`):

```yaml
# Database connection URL
url: ${DRIFT_URL:-postgres://localhost:5432/mydb?sslmode=disable}

# Migration file locations (relative to drift.yaml)
locations:
  - migrations

# Schema history table name
table: drift_schema_history

# Schemas to manage (Postgres only)
schemas:
  - public

# Baseline version (skip migrations at or below this version)
# baseline: "001"

# Allow out-of-order migrations
# outOfOrder: false

# Placeholder substitution in migration SQL
# placeholders:
#   schema: public
#   table_prefix: app_

# Callback locations
# callbacks:
#   - callbacks

# Web UI settings
ui:
  port: 4077
  host: 127.0.0.1
```

## Environment Variable Interpolation

Use `${VAR}` syntax in your config file. Default values are supported with `${VAR:-default}`:

```yaml
url: ${DATABASE_URL:-postgres://localhost:5432/mydb}
```

This reads `DATABASE_URL` from the environment, falling back to the default if unset.

## Settings Reference

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `url` | string | — | Database connection URL (required) |
| `locations` | string[] | `["migrations"]` | Directories containing migration files |
| `table` | string | `"drift_schema_history"` | Name of the history tracking table |
| `schemas` | string[] | `["public"]` | Schemas to manage (Postgres only) |
| `baseline` | string | — | Skip migrations at or below this version |
| `outOfOrder` | bool | `false` | Allow applying migrations out of sequence |
| `placeholders` | map | — | Key-value pairs for SQL placeholder substitution |
| `callbacks` | string[] | — | Directories containing callback SQL files |
| `ui.port` | int | `4077` | Web UI port |
| `ui.host` | string | `"127.0.0.1"` | Web UI bind address |

## CLI Flag Overrides

CLI flags take precedence over `drift.yaml`:

```bash
# Override URL
drift migrate --url postgres://prod-host:5432/mydb

# Override config file location
drift migrate -c /path/to/drift.yaml

# Override migration locations
drift migrate --locations migrations,seeds

# Override history table name
drift migrate --table my_schema_history
```

## Global Flags

These flags work with all commands:

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | `-c` | Path to config file (default: `drift.yaml`) |
| `--url` | — | Database connection URL |
| `--locations` | — | Migration file locations (comma-separated) |
| `--table` | — | Schema history table name |
| `--schemas` | — | Schemas to manage (comma-separated) |
| `--verbose` | `-v` | Enable verbose output |
| `--quiet` | `-q` | Suppress non-essential output |

## URL Formats

Drift auto-detects the database driver from the URL scheme:

```bash
# PostgreSQL
postgres://user:pass@host:5432/dbname?sslmode=disable
postgresql://user:pass@host:5432/dbname

# MySQL
mysql://user:pass@host:3306/dbname

# SQLite (file path or scheme)
sqlite:///path/to/db.sqlite
/path/to/db.sqlite
./local.db
```
