# Postgres Example

A hands-on walkthrough of drift with PostgreSQL. You'll apply migrations, inspect schema, take snapshots, undo changes, and more.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [drift](https://github.com/dunkinfrunkin/drift)

## 1. Start Postgres

```bash
make up
```

Or manually:

```bash
docker compose up -d
```

The database starts with some pre-existing tables (`products`, `orders`) and seed data. Drift only manages migrations you tell it about — these tables are left untouched.

## 2. Initialize drift

```bash
drift init --driver postgres
```

Update the `url` in `drift.yaml`:

```yaml
url: postgres://drift:drift@localhost:5432/drift_example?sslmode=disable
```

## 3. Run Migrations

```bash
drift migrate
```

This applies all pending versioned migrations (V001 through V005) and the repeatable migration (R__create_views.sql).

## 4. Explore

Check what was applied:

```bash
drift info
```

Validate that applied migrations match the local files:

```bash
drift validate
```

## 5. Schema Snapshots

Capture the current database schema:

```bash
drift snapshot -o snapshot.json
```

Compare the live database against the snapshot (should show no differences):

```bash
drift diff snapshot.json
```

Generate SQL output of the full schema:

```bash
drift diff --format sql
```

## 6. Undo Migrations

Undo scripts are stored as `.bak` files. Enable them with:

```bash
make enable-undo
```

Undo the last migration:

```bash
drift undo
```

Undo multiple migrations:

```bash
drift undo --count 2
```

Check the state after undoing:

```bash
drift info
```

Re-apply everything:

```bash
drift migrate
```

When you're done, disable the undo scripts:

```bash
make disable-undo
```

## 7. Dry Run

Preview what would be applied without making changes:

```bash
drift migrate --dry-run
```

## 8. Linting

Check migration files for common issues:

```bash
drift lint
```

## 9. Web Dashboard

Launch the built-in web UI:

```bash
drift ui
```

Opens a dashboard at `http://localhost:4077` with migration history, schema diff, lint results, and more.

## 10. Teardown

```bash
make down
```

Clean up generated files:

```bash
make clean
```

## What's in the Box

```
seed/
  init.sql                          # pre-existing tables (products, orders) drift won't touch

migrations/
  V001__create_users.sql            # users table
  V002__create_posts.sql            # posts table with FK to users
  V003__add_user_status.sql         # adds status column to users
  V004__create_comments.sql         # comments table with FKs
  V005__add_indexes.sql             # indexes for foreign keys
  R__create_views.sql               # repeatable: active_users view
  U001__drop_users.sql.bak          # undo for V001 (enable with make enable-undo)
  U002__drop_posts.sql.bak          # undo for V002
  U003__remove_user_status.sql.bak  # undo for V003
  U004__drop_comments.sql.bak       # undo for V004
  U005__drop_indexes.sql.bak        # undo for V005
```
