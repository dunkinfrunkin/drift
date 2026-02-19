---
title: drift validate
description: Validate applied migrations against local files.
---

# drift validate

Validate that applied migrations match the local migration files. Catches checksum mismatches, missing files, and failed entries.

## Usage

```bash
drift validate
```

## What It Checks

- **Checksum mismatches** — A migration file was modified after being applied
- **Missing files** — A migration was applied but the file no longer exists
- **Failed entries** — A migration was attempted but failed

## Examples

```bash
# Validate migrations
drift validate
```

Successful output:
```
Validation passed: all migrations are consistent.
```

Failed output:
```
Validation failed:
  - Checksum mismatch: V002 (expected a1b2c3d4, got e5f6a7b8)
  - Missing migration file: V003
```

:::tip
Run `drift repair` to fix checksum mismatches by updating stored checksums to match current files.
:::
