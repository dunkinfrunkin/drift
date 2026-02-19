---
title: drift migrate
description: Apply pending migrations to the database.
---

# drift migrate

Apply pending database migrations in version order.

## Usage

```bash
drift migrate [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview SQL without executing |
| `--target VERSION` | Migrate up to this version (e.g., `V005`) |
| `--cherry-pick VERSIONS` | Apply only specific versions (comma-separated, e.g., `V003,V005`) |
| `--skip VERSIONS` | Skip specific versions (comma-separated) |
| `--out-of-order` | Allow applying migrations older than the latest applied |

## Examples

```bash
# Apply all pending migrations
drift migrate

# Preview what would run
drift migrate --dry-run

# Migrate up to V005 only
drift migrate --target V005

# Apply only specific versions
drift migrate --cherry-pick V003,V005

# Skip a problematic migration
drift migrate --skip V004

# Allow out-of-order application
drift migrate --out-of-order
```

## Behavior

1. Acquires an advisory lock to prevent concurrent migrations
2. Discovers migration files from configured locations
3. Compares against the schema history table
4. Plans the migration (respecting target, skip, cherry-pick flags)
5. Executes each pending migration in a transaction (if the database supports transactional DDL)
6. Records each successful migration in the history table
7. Releases the advisory lock

:::info
PostgreSQL and SQLite support transactional DDL — if a migration fails, changes are rolled back. MySQL does **not** support transactional DDL for schema changes.
:::
