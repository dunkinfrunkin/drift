---
title: drift info
description: Display migration status information.
---

# drift info

Show the status of all discovered and applied migrations.

## Usage

```bash
drift info [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--format FORMAT` | Output format: `table` (default), `json`, `yaml` |

## Examples

```bash
# Show migration status as a table
drift info

# Output as JSON (useful for scripting)
drift info --format json

# Output as YAML
drift info --format yaml
```

## Output

The table view shows each migration's version, state, description, checksum, and installation time:

```
+-----------+---------+-------------------+----------+---------------------+
|  Version  |  State  |    Description    | Checksum |     Installed On    |
+-----------+---------+-------------------+----------+---------------------+
|  V001     | Applied | create_users      | a1b2c3d4 | 2025-01-15 10:30:00 |
|  V002     | Applied | add_email_index   | e5f6a7b8 | 2025-01-15 10:30:01 |
|  V003     | Pending | create_orders     |          |                     |
+-----------+---------+-------------------+----------+---------------------+
```

**States:**
- **Applied** — Migration has been successfully run
- **Pending** — Migration file exists but hasn't been applied
- **Failed** — Migration was attempted but failed
