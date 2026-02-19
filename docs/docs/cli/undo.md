---
title: drift undo
description: Rollback applied migrations.
---

# drift undo

Undo previously applied migrations using undo scripts (`U{version}__*.sql`).

## Usage

```bash
drift undo [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--count N` | Number of migrations to undo (default: 1) |
| `--target VERSION` | Undo down to this version |
| `--dry-run` | Preview SQL without executing |

## Examples

```bash
# Undo the last migration
drift undo

# Undo the last 3 migrations
drift undo --count 3

# Undo down to V002
drift undo --target V002

# Preview what would be undone
drift undo --dry-run
```

## Behavior

Migrations are undone in reverse order (most recent first). Each undo migration runs the corresponding `U{version}__*.sql` file and removes the entry from the history table.

:::warning
Undo requires corresponding undo files (`U001__*.sql`, etc.) to exist in your migration locations. If an undo file is missing, the operation will fail.
:::
