---
sidebar_position: 4
title: Migration Files
description: Naming conventions for versioned, undo, repeatable, and callback migrations.
---

# Migration Files

Drift discovers migration files by naming convention. Place them in your configured `locations` directory (default: `migrations/`).

## Versioned Migrations

Standard forward migrations. Applied once, in order.

```
V001__create_users.sql
V002__add_email_index.sql
V003__create_orders.sql
```

**Pattern:** `V{version}__{description}.sql`

- **Version**: Numeric identifier (e.g., `001`, `002`, `1`, `2`, `100`)
- **Separator**: Double underscore `__`
- **Description**: Human-readable name (underscores become spaces in output)

## Undo Migrations

Rollback scripts for versioned migrations. Used by `drift undo`.

```
U001__drop_users.sql
U002__remove_email_index.sql
U003__drop_orders.sql
```

**Pattern:** `U{version}__{description}.sql`

Each undo file should reverse the changes made by the corresponding versioned migration.

:::tip
Always write undo migrations alongside your forward migrations. This makes rollbacks reliable and testable.
:::

### Example Pair

**`V001__create_users.sql`**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL
);
```

**`U001__drop_users.sql`**
```sql
DROP TABLE IF EXISTS users;
```

## Repeatable Migrations

Applied every time their content changes. Useful for views, functions, and stored procedures.

```
R__create_user_view.sql
R__refresh_materialized_views.sql
```

**Pattern:** `R__{description}.sql`

Repeatable migrations:
- Run **after** all versioned migrations
- Re-run whenever their checksum changes
- Have no version number

## Callbacks

SQL files that run at specific points in the migration lifecycle.

```
beforeMigrate__audit_log.sql
afterMigrate__notify_complete.sql
beforeEachMigrate__set_role.sql
afterEachMigrate__log_step.sql
```

**Supported callbacks:**

| Callback | When it runs |
|----------|-------------|
| `beforeMigrate` | Before the migration run starts |
| `afterMigrate` | After all migrations complete |
| `beforeEachMigrate` | Before each individual migration |
| `afterEachMigrate` | After each individual migration |

Place callback files in your `callbacks` directory (configured in `drift.yaml`).

## Placeholder Substitution

Use `${placeholder}` syntax in your SQL files:

```sql
CREATE TABLE ${schema}.${table_prefix}users (
    id SERIAL PRIMARY KEY
);
```

Configure placeholders in `drift.yaml`:

```yaml
placeholders:
  schema: public
  table_prefix: app_
```

## Checksums

Drift computes a CRC32 checksum for each migration file. After a migration is applied, drift validates that the file hasn't been modified. If a checksum mismatch is detected:

- `drift validate` will report the mismatch
- `drift repair` will update the stored checksum to match the current file

## Directory Structure Example

```
migrations/
├── V001__create_users.sql
├── V002__add_email_index.sql
├── V003__create_orders.sql
├── U001__drop_users.sql
├── U002__remove_email_index.sql
├── U003__drop_orders.sql
├── R__user_statistics_view.sql
└── R__order_summary_view.sql
callbacks/
├── beforeMigrate__set_search_path.sql
└── afterMigrate__refresh_cache.sql
```
