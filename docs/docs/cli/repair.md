---
title: drift repair
description: Repair the schema history table.
---

# drift repair

Fix the schema history table by removing failed entries and realigning checksums with current migration files.

## Usage

```bash
drift repair
```

## What It Does

1. **Removes failed entries** — Deletes history rows where `success = false`
2. **Realigns checksums** — Updates stored checksums to match current file contents

## When to Use

- After a migration fails and you've fixed the SQL file
- After intentionally modifying an already-applied migration file
- When `drift validate` reports checksum mismatches

## Example

```bash
# Validate shows issues
drift validate
# Validation failed:
#   - Checksum mismatch: V002
#   - Failed migration: V003

# Repair fixes them
drift repair
# Repaired: removed 1 failed entry, updated 1 checksum
```
