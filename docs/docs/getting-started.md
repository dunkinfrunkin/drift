---
sidebar_position: 1
title: Getting Started
description: Get up and running with drift in under a minute.
---

# Getting Started

Drift is a fast, open-source database migration tool. Single binary, zero runtime dependencies. Everything Flyway paywalls — undo, dry-run, cherry-pick, drift detection, schema diff, linting — is free and open source.

## Quick Start

### 1. Install drift

```bash
brew install dunkinfrunkin/tap/drift
```

See [Installation](./installation.md) for other methods (curl, Docker, Go).

### 2. Initialize a project

```bash
drift init --driver postgres
```

This creates a `drift.yaml` config file and a `migrations/` directory.

### 3. Create your first migration

Create a file `migrations/V001__create_users.sql`:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 4. Run migrations

```bash
drift migrate --url postgres://localhost:5432/mydb
```

### 5. Check status

```bash
drift info --url postgres://localhost:5432/mydb
```

```
+-----------+---------+-------------------+---------+---------------------+
|  Version  |  State  |    Description    | Checksum|     Installed On    |
+-----------+---------+-------------------+---------+---------------------+
|  V001     | Applied | create_users      | a1b2c3d4| 2025-01-15 10:30:00 |
+-----------+---------+-------------------+---------+---------------------+
```

### 6. Undo if needed

```bash
drift undo --url postgres://localhost:5432/mydb
```

### 7. Launch the web UI

```bash
drift ui --url postgres://localhost:5432/mydb
```

Opens an embedded dashboard at `http://localhost:4077` with migration history, schema diffs, lint results, and more.

## Try It with Docker

Want to try drift hands-on? The [`example/postgres`](https://github.com/dunkinfrunkin/drift/tree/main/example/postgres) directory has a Docker Compose setup with sample migrations, undo scripts, and a step-by-step walkthrough covering migrate, info, snapshot, diff, undo, lint, and the web UI.

## What's Next?

- **[Migration Files](./migration-files.md)** — Naming conventions, undo files, repeatable migrations
- **[Configuration](./configuration.md)** — YAML config with environment variable support
- **[CLI Reference](./cli-reference.md)** — All commands and flags
- **[Database Support](./databases.md)** — PostgreSQL, MySQL, SQLite details
