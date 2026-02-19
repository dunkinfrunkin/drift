---
title: drift clean
description: Drop the schema history table.
---

# drift clean

Drop the drift schema history table. This removes all migration tracking data.

## Usage

```bash
drift clean [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--yes` | Skip confirmation prompt |

## Examples

```bash
# Clean with confirmation prompt
drift clean

# Skip confirmation
drift clean --yes
```

:::danger
This is a destructive operation. It drops the history table, not your application tables. After cleaning, drift will treat all migrations as pending. Use with caution, primarily in development.
:::
