---
title: drift diff
description: Compare database schemas.
---

# drift diff

Compare the current database schema against a snapshot or show the full schema as a diff.

## Usage

```bash
drift diff [snapshot-file] [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--format FORMAT` | Output format: `text` (default), `json`, `sql` |

## Examples

```bash
# Show current schema diff (against empty)
drift diff

# Compare against a snapshot
drift diff schema-snapshot.json

# Output as SQL statements
drift diff --format sql

# Output as JSON
drift diff schema-snapshot.json --format json
```

## Output

Text format shows added, removed, and modified tables, columns, and indexes:

```
+ Table: orders
    + Column: id (integer, NOT NULL)
    + Column: user_id (integer, NOT NULL)
    + Column: total (numeric(10,2))
    + Index: idx_orders_user_id (user_id)

~ Table: users
    + Column: updated_at (timestamp)
    ~ Column: email (varchar(255) -> varchar(512))
```

The `sql` format generates the DDL statements needed to transform one schema into the other.
