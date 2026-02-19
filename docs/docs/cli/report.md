---
title: drift report
description: Generate a migration report.
---

# drift report

Generate a report of all migrations with their status, execution times, and history.

## Usage

```bash
drift report [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--format FORMAT` | Output format: `text` (default), `json`, `html` |

## Examples

```bash
# Text report
drift report

# JSON for programmatic use
drift report --format json

# HTML report (styled, dark theme)
drift report --format html > report.html
```

The HTML format generates a self-contained, styled report suitable for sharing or archiving.
