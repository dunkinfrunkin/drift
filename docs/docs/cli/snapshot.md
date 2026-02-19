---
title: drift snapshot
description: Capture a schema snapshot.
---

# drift snapshot

Capture the current database schema as a snapshot file. Useful for tracking schema changes over time and for use with `drift diff`.

## Usage

```bash
drift snapshot [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--output FILE` | Write snapshot to file (default: stdout) |
| `--format FORMAT` | Output format: `json` (default), `yaml` |

## Examples

```bash
# Print snapshot to stdout
drift snapshot

# Save to file
drift snapshot --output schema.json

# Save as YAML
drift snapshot --format yaml --output schema.yaml
```

## Snapshot Contents

A snapshot captures all tables, columns, indexes, and foreign keys:

```json
{
  "tables": [
    {
      "name": "users",
      "columns": [
        {"name": "id", "type": "integer", "nullable": false, "primaryKey": true},
        {"name": "email", "type": "varchar(255)", "nullable": false}
      ],
      "indexes": [
        {"name": "users_email_key", "columns": ["email"], "unique": true}
      ]
    }
  ]
}
```
