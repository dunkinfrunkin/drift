---
sidebar_position: 7
title: Linting
description: Built-in SQL lint rules for migration safety.
---

# Linting

Drift includes a built-in SQL linter that catches dangerous patterns and enforces best practices in your migration files.

## Running the Linter

```bash
# Lint all migration files
drift lint

# Run specific rules only
drift lint --rules no-drop-table,naming-convention
```

## Built-in Rules

### no-drop-table

**Severity:** Error

Detects `DROP TABLE` statements. Dropping tables in production can cause irreversible data loss.

```sql
-- Flagged
DROP TABLE users;
DROP TABLE IF EXISTS legacy_data;
```

### no-drop-column

**Severity:** Error

Detects `DROP COLUMN` statements. Dropping columns causes data loss and can break running applications.

```sql
-- Flagged
ALTER TABLE users DROP COLUMN legacy_field;
```

### no-raw-truncate

**Severity:** Error

Detects `TRUNCATE` statements. Truncating tables deletes all data without row-level logging.

```sql
-- Flagged
TRUNCATE TABLE sessions;
```

### require-transaction-safe

**Severity:** Warning

Flags operations that aren't safe within transactions, such as `CREATE INDEX CONCURRENTLY` in PostgreSQL.

### naming-convention

**Severity:** Warning

Checks that migration filenames follow the expected naming convention:
- Versioned: `V{number}__{description}.sql`
- Undo: `U{number}__{description}.sql`
- Repeatable: `R__{description}.sql`

### no-select-star

**Severity:** Warning

Detects `SELECT *` statements. While sometimes acceptable in migrations, explicit column lists are preferred.

```sql
-- Flagged
INSERT INTO new_users SELECT * FROM old_users;

-- Preferred
INSERT INTO new_users (id, email, name) SELECT id, email, name FROM old_users;
```

## Output Format

```
V003__drop_legacy.sql:
  [ERROR] no-drop-table: DROP TABLE detected — use caution in production
  [WARN]  no-select-star: SELECT * detected — specify columns explicitly

V005__update_users.sql:
  [ERROR] no-drop-column: DROP COLUMN detected — may cause data loss

2 files scanned, 3 issues found (2 errors, 1 warning)
```

## Filtering Rules

Run only the rules you care about:

```bash
# Only check for destructive operations
drift lint --rules no-drop-table,no-drop-column,no-raw-truncate

# Only check naming
drift lint --rules naming-convention
```
