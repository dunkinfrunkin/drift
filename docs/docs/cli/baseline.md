---
title: drift baseline
description: Baseline an existing database.
---

# drift baseline

Mark an existing database as having a baseline version. All migrations up to and including the baseline version will be skipped.

## Usage

```bash
drift baseline [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--version VERSION` | Version to baseline at (e.g., `V001`) |
| `--from-db` | Generate a baseline migration from the current database schema |
| `--output FILE` | Write generated SQL to file (used with `--from-db`) |

## Examples

```bash
# Baseline at version V005
drift baseline --version V005

# Generate a baseline SQL file from current schema
drift baseline --from-db --output baseline.sql
```

## When to Use

**Adopting drift on an existing database:**

If your database already has tables and you're starting to use drift, baseline tells drift to skip migrations that represent the existing schema:

```bash
# 1. Create migrations matching your existing schema
# 2. Baseline so drift knows they're already applied
drift baseline --version V010

# 3. Future migrations will start from V011
drift migrate
```

**Generating a baseline file:**

```bash
drift baseline --from-db --output migrations/V000__baseline.sql
```

This introspects the live database and generates `CREATE TABLE` statements for all existing tables.
