---
title: drift ui
description: Launch the embedded web dashboard.
---

# drift ui

Start the embedded web dashboard for visualizing migration state, schema diffs, lint results, and more.

## Usage

```bash
drift ui [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--port PORT` | Port to listen on (default: `4077`) |
| `--host HOST` | Host to bind to (default: `127.0.0.1`) |
| `--no-open` | Don't auto-open browser |

## Examples

```bash
# Start on default port
drift ui

# Custom port
drift ui --port 8080

# Bind to all interfaces
drift ui --host 0.0.0.0

# Don't open browser automatically
drift ui --no-open
```

## Dashboard Pages

The web UI includes:

- **Dashboard** — Overview with migration stats, validation status, and recent activity
- **Migrations** — Full migration list with timeline view
- **Schema Diff** — Visual schema comparison with change badges
- **Schema Browser** — Table/column/index explorer with PK/FK indicators
- **Lint Results** — Severity-tagged issue list for all migration files
- **Actions** — Run migrate, undo, repair, and other operations from the browser

See [Web UI](../web-ui.md) for more details.
