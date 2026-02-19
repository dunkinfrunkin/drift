---
sidebar_position: 6
title: Database Support
description: PostgreSQL, MySQL, and SQLite support details.
---

# Database Support

Drift supports three databases out of the box. The database driver is auto-detected from the connection URL scheme.

## PostgreSQL

**Driver:** [pgx/v5](https://github.com/jackc/pgx)

```yaml
url: postgres://user:password@localhost:5432/dbname?sslmode=disable
```

| Feature | Support |
|---------|---------|
| Transactional DDL | Yes |
| Advisory locking | `pg_advisory_lock` |
| Schema introspection | `information_schema` + `pg_indexes` |
| Multiple schemas | Yes (`--schemas` flag) |
| Placeholder style | `$1, $2, $3` |

### URL Formats

```
postgres://user:pass@host:5432/dbname
postgres://user:pass@host:5432/dbname?sslmode=require
postgresql://user:pass@host:5432/dbname
```

## MySQL

**Driver:** [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

```yaml
url: mysql://user:password@localhost:3306/dbname
```

| Feature | Support |
|---------|---------|
| Transactional DDL | No |
| Advisory locking | `GET_LOCK` / `RELEASE_LOCK` |
| Schema introspection | `information_schema` |
| Multiple schemas | No |
| Placeholder style | `?` |

:::warning
MySQL does not support transactional DDL. If a migration fails partway through, the already-executed statements will **not** be rolled back. Write defensive migrations for MySQL.
:::

### URL Format

```
mysql://user:pass@host:3306/dbname
```

## SQLite

**Driver:** [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO)

```yaml
url: sqlite:///path/to/database.db
```

| Feature | Support |
|---------|---------|
| Transactional DDL | Yes |
| Advisory locking | In-process mutex |
| Schema introspection | `sqlite_master` + `PRAGMA table_info` |
| Multiple schemas | No |
| Placeholder style | `?` |

SQLite uses a pure Go driver — no C compiler or CGO required. The single-binary promise is maintained.

### URL Formats

```
sqlite:///path/to/db.sqlite
sqlite:///absolute/path/to/db.sqlite
./local.db
/path/to/database.sqlite3
```

Any path ending in `.db`, `.sqlite`, or `.sqlite3` is auto-detected as SQLite.

### WAL Mode

Drift enables WAL (Write-Ahead Logging) mode for SQLite databases, which improves concurrent read performance.

## Advisory Locking

Drift uses advisory locks to prevent concurrent migration runs:

| Database | Lock mechanism | Scope |
|----------|---------------|-------|
| PostgreSQL | `pg_advisory_lock(2020817574)` | Database-wide |
| MySQL | `GET_LOCK('drift_migration', 10)` | Instance-wide |
| SQLite | In-process `sync.Mutex` | Process-wide |

The lock is acquired before any migration operation and released when complete.

## Schema History Table

All databases use the same history table structure (`drift_schema_history` by default):

| Column | Type | Description |
|--------|------|-------------|
| `installed_rank` | integer (PK) | Auto-incrementing order |
| `version` | varchar | Migration version |
| `description` | varchar | Migration description |
| `type` | varchar | `SQL`, `UNDO`, `BASELINE`, `REPEATABLE` |
| `script` | varchar | Migration filename |
| `checksum` | integer | CRC32 checksum |
| `installed_by` | varchar | User who ran the migration |
| `installed_on` | timestamp | When it was applied |
| `execution_time` | integer | Duration in milliseconds |
| `success` | boolean | Whether it succeeded |
