# drift

A fast, open-source database migration tool. Single binary, zero runtime dependencies.

**Everything Flyway paywalls — undo, dry-run, cherry-pick, drift detection, schema diff, linting — is free and open source.**

## Features

- **Migrate** — Apply versioned and repeatable migrations
- **Undo** — Reverse migrations with undo scripts
- **Dry-run** — Preview what would change without executing
- **Cherry-pick** — Apply specific versions out of order
- **Schema diff** — Compare database states, detect drift
- **Snapshot** — Capture and serialize schema state
- **Lint** — Catch dangerous SQL patterns before they run
- **Squash** — Consolidate migrations into a single file
- **Web UI** — Embedded dashboard, no extra dependencies
- **3 databases** — PostgreSQL, MySQL, SQLite (pure Go, no CGO)

## Install

```bash
# Homebrew
brew install frankchan/tap/drift

# curl
curl -fsSL https://raw.githubusercontent.com/frankchan/drift/main/install.sh | sh

# Go
go install github.com/frankchan/drift/cmd/drift@latest

# Docker
docker run --rm ghcr.io/frankchan/drift version
```

## Quick Start

```bash
# Initialize a project
drift init --driver postgres

# Edit drift.yaml with your database URL
# Create migration files in migrations/

# Apply migrations
drift migrate

# Check status
drift info

# Undo the last migration
drift undo

# Launch the web UI
drift ui
```

## Migration Files

```
migrations/
├── V001__create_users.sql       # Versioned (applied once, in order)
├── V002__add_email_index.sql
├── V003__create_posts.sql
├── U001__drop_users.sql          # Undo (reverses V001)
├── U002__drop_email_index.sql
├── R__refresh_views.sql          # Repeatable (re-applied when changed)
├── beforeMigrate__audit.sql      # Callback (lifecycle hooks)
└── afterMigrate__notify.sql
```

## Configuration

`drift.yaml`:

```yaml
url: postgres://user:pass@localhost:5432/mydb?sslmode=disable
locations:
  - migrations
table: drift_schema_history
# placeholders:
#   schema: public
```

Environment variables are supported with `${VAR}` and `${VAR:-default}` syntax.

## CLI Reference

```
drift migrate     [--dry-run] [--target V005] [--cherry-pick V003,V005] [--skip V003] [--out-of-order]
drift undo        [--count 1] [--target V003] [--dry-run]
drift info        [--format table|json|yaml]
drift validate
drift clean       [--yes]
drift baseline    [--version V001] [--from-db] [--output file.sql]
drift repair
drift diff        [source] [--format text|json|sql]
drift snapshot    [--output file] [--format json|yaml]
drift report      [--format text|html|json]
drift lint        [--rules ...]
drift squash      [--up-to V010] [--output file.sql]
drift ui          [--port 4077] [--host 127.0.0.1] [--no-open]
drift init        [--driver postgres]
drift version
```

**Global flags:** `--config/-c`, `--url`, `--locations`, `--table`, `--schemas`, `--verbose/-v`, `--quiet/-q`

## Supported Databases

| Database   | Driver                | URL Format                                  | Advisory Lock      |
|------------|----------------------|---------------------------------------------|--------------------|
| PostgreSQL | `pgx` (pure Go)     | `postgres://user:pass@host:5432/db`         | `pg_advisory_lock` |
| MySQL      | `go-sql-driver`      | `mysql://user:pass@tcp(host:3306)/db`       | `GET_LOCK`         |
| SQLite     | `modernc.org/sqlite` | `sqlite://path.db` or `data.db`             | In-process mutex   |

## Lint Rules

| Rule                      | Severity | Description                                        |
|---------------------------|----------|----------------------------------------------------|
| `no-drop-table`           | ERROR    | DROP TABLE without IF EXISTS                       |
| `no-drop-column`          | WARNING  | DROP COLUMN detected                               |
| `no-raw-truncate`         | WARNING  | TRUNCATE bypasses triggers                         |
| `require-transaction-safe`| WARNING  | Operations that can't run in transactions          |
| `naming-convention`       | WARNING  | Missing description in filename                    |
| `no-select-star`          | WARNING  | SELECT * in migrations                             |

## Web UI

Run `drift ui` to start the embedded web dashboard at `http://localhost:4077`.

Pages: Dashboard, Migration List with Timeline, Schema Diff Viewer, Schema Browser, Lint Results, Actions (migrate/undo/repair from the browser).

## Building from Source

```bash
git clone https://github.com/frankchan/drift.git
cd drift
make build    # builds UI + Go binary → bin/drift
make test     # runs Go tests
```

Requires: Go 1.23+, Node.js 20+

## License

MIT
