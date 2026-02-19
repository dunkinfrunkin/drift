---
sidebar_position: 5
title: CLI Reference
description: Complete reference for all drift commands and flags.
---

# CLI Reference

Drift provides 15 commands covering the full migration lifecycle. Every feature is available from the CLI — no paid tiers, no feature gates.

## Command Overview

| Command | Description |
|---------|-------------|
| [`drift migrate`](./cli/migrate.md) | Apply pending migrations |
| [`drift undo`](./cli/undo.md) | Rollback applied migrations |
| [`drift info`](./cli/info.md) | Show migration status |
| [`drift validate`](./cli/validate.md) | Validate applied migrations |
| [`drift diff`](./cli/diff.md) | Compare database schemas |
| [`drift snapshot`](./cli/snapshot.md) | Capture schema snapshot |
| [`drift lint`](./cli/lint.md) | Lint migration files |
| [`drift baseline`](./cli/baseline.md) | Baseline an existing database |
| [`drift repair`](./cli/repair.md) | Repair schema history |
| [`drift clean`](./cli/clean.md) | Drop schema history table |
| [`drift squash`](./cli/squash.md) | Squash migrations |
| [`drift report`](./cli/report.md) | Generate migration report |
| [`drift ui`](./cli/ui.md) | Launch web dashboard |
| [`drift init`](./cli/init.md) | Initialize new project |
| `drift version` | Show version info |

## Global Flags

Available on all commands:

```
--config, -c    Path to config file (default: drift.yaml)
--url           Database connection URL
--locations     Migration file locations (comma-separated)
--table         Schema history table name
--schemas       Schemas to manage (comma-separated)
--verbose, -v   Enable verbose output
--quiet, -q     Suppress non-essential output
```
