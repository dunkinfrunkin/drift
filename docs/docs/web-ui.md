---
sidebar_position: 8
title: Web UI
description: Embedded web dashboard for migration management.
---

# Web UI

Drift includes an embedded web dashboard — no separate install, no Docker compose, no npm. It's built into the binary.

## Starting the UI

```bash
drift ui --url postgres://localhost:5432/mydb
```

Opens `http://localhost:4077` in your browser.

### Options

```bash
# Custom port
drift ui --port 8080

# Bind to all interfaces (for remote access)
drift ui --host 0.0.0.0

# Don't auto-open browser
drift ui --no-open
```

## Pages

### Dashboard

Overview of your migration state:
- Total migrations (applied, pending, failed)
- Validation status
- Lint issue count
- Current configuration
- Recent migration activity

### Migrations

Full migration list with:
- Version, state, description, checksum
- Execution time for each migration
- Timeline view showing migration history

### Schema Diff

Visual schema comparison:
- Added tables/columns highlighted in green
- Removed tables/columns highlighted in red
- Modified columns with before/after comparison
- SQL preview of diff

### Schema Browser

Interactive schema explorer:
- Browse all tables and their columns
- See column types, nullability, defaults
- Primary key and foreign key indicators
- Index listing per table

### Lint Results

Migration file analysis:
- Issues grouped by file
- Severity badges (error, warning)
- Rule name and description for each issue

### Actions

Run operations directly from the browser:
- Migrate, undo, repair, clean, baseline
- Action log showing results
- Real-time feedback via Server-Sent Events

## API

The web UI is backed by a REST API at `/api/v1/`:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/info` | GET | Migration info |
| `/api/v1/validate` | GET | Validation results |
| `/api/v1/migrate` | POST | Run migrations |
| `/api/v1/undo` | POST | Undo migrations |
| `/api/v1/repair` | POST | Repair history |
| `/api/v1/clean` | POST | Clean history |
| `/api/v1/baseline` | POST | Set baseline |
| `/api/v1/diff` | GET | Schema diff |
| `/api/v1/snapshot` | GET | Schema snapshot |
| `/api/v1/lint` | GET | Lint results |
| `/api/v1/config` | GET | Current config |
| `/api/v1/events` | GET | SSE event stream |

## Architecture

The web UI is a React + TypeScript SPA built with Vite and Tailwind CSS. It's compiled to static files and embedded into the Go binary via `go:embed`. No Node.js runtime is needed at deployment time.
