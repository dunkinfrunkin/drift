---
title: drift lint
description: Lint migration files for common issues.
---

# drift lint

Scan migration files for dangerous patterns, naming violations, and best practice issues.

## Usage

```bash
drift lint [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--rules RULES` | Only run specific rules (comma-separated) |

## Examples

```bash
# Lint all migration files
drift lint

# Run only specific rules
drift lint --rules no-drop-table,naming-convention
```

## Output

```
V003__drop_legacy.sql:
  [ERROR] no-drop-table: DROP TABLE detected — use caution in production
  [WARN]  no-select-star: SELECT * detected — specify columns explicitly

V005__update_users.sql:
  [ERROR] no-drop-column: DROP COLUMN detected — may cause data loss
```

See [Linting](../linting.md) for the full list of built-in rules and how to configure them.
