#!/usr/bin/env bash
set -e

# Colors
BOLD='\033[1m'
DIM='\033[2m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
RESET='\033[0m'

step=0

banner() {
  step=$((step + 1))
  echo ""
  echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
  echo -e "${BOLD}  Step $step: $1${RESET}"
  echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
  echo ""
}

info() {
  echo -e "${DIM}$1${RESET}"
}

run() {
  echo -e "${YELLOW}\$ $*${RESET}"
  "$@"
  echo ""
}

pause() {
  echo ""
  echo -e "${GREEN}Next up: $1${RESET}"
  echo -e "${GREEN}Press Enter to continue...${RESET}"
  read -r
}

cd "$(dirname "$0")"

# ─── Dependency check ─────────────────────────────────────────────────────────

RED='\033[0;31m'
missing=0

if ! command -v docker &> /dev/null; then
  echo -e "${RED}Error: docker is not installed.${RESET}"
  echo -e "${DIM}Install it from https://docs.docker.com/get-docker/${RESET}"
  missing=1
fi

if ! command -v drift &> /dev/null; then
  echo -e "${RED}Error: drift is not installed.${RESET}"
  echo -e "${DIM}Install it with: brew install dunkinfrunkin/tap/drift${RESET}"
  missing=1
fi

if [ "$missing" -eq 1 ]; then
  echo ""
  echo -e "${RED}Please install the missing dependencies and try again.${RESET}"
  exit 1
fi

if ! docker info &> /dev/null; then
  echo -e "${RED}Error: Docker daemon is not running.${RESET}"
  echo -e "${DIM}Start Docker Desktop or run 'dockerd' and try again.${RESET}"
  exit 1
fi

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════════╗${RESET}"
echo -e "${BOLD}${CYAN}║           drift — Interactive Tutorial               ║${RESET}"
echo -e "${BOLD}${CYAN}║                                                      ║${RESET}"
echo -e "${BOLD}${CYAN}║  Learn drift by running real commands against a      ║${RESET}"
echo -e "${BOLD}${CYAN}║  local Postgres database. Press Enter at each step.  ║${RESET}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════════╝${RESET}"
echo ""
echo -e "${GREEN}Press Enter to start...${RESET}"
read -r

# ─── Reset ───────────────────────────────────────────────────────────────────

info "Resetting environment..."
docker compose down -v 2>/dev/null || true
rm -f drift.yaml snapshot.json migrations/U*.sql
echo -e "${GREEN}Clean slate.${RESET}"
echo ""

# ─── Step 1: Start Postgres ──────────────────────────────────────────────────

banner "Start Postgres"
info "Spinning up a Postgres 16 container with Docker Compose."
info "The database comes pre-loaded with some tables (products, orders)"
info "that drift won't touch — only your migrations are managed."
echo ""
run docker compose up -d
echo "Waiting for Postgres to be ready..."
until docker compose exec postgres pg_isready -U drift -d drift_example -p 5432 > /dev/null 2>&1; do
  sleep 1
done
echo -e "${GREEN}Postgres is ready.${RESET}"
pause "Initialize a drift project"

# ─── Step 2: Initialize drift ────────────────────────────────────────────────

banner "Initialize drift"
info "This creates a drift.yaml config file and a migrations/ directory."
info "We'll then update the database URL to point to our local Postgres."
echo ""
rm -f drift.yaml
run drift init --driver postgres

# Patch the URL
sed -i.bak 's|postgres://user:password@localhost:5432/mydb?sslmode=disable|postgres://drift:drift@localhost:5433/drift_example?sslmode=disable|' drift.yaml
rm -f drift.yaml.bak
echo -e "${DIM}Updated drift.yaml with the correct database URL.${RESET}"
echo ""
echo -e "${DIM}Contents of drift.yaml:${RESET}"
cat drift.yaml
pause "Capture a baseline schema snapshot"

# ─── Step 3: Schema snapshot (before migrations) ────────────────────────────

banner "Schema Snapshot (before migrations)"
info "Capture the current database schema to a JSON file BEFORE applying"
info "any migrations. This gives us a baseline to compare against later."
info "Right now the database only has the pre-existing seed tables (products, orders)."
echo ""
run drift snapshot -o snapshot.json
echo -e "${DIM}Snapshot saved to snapshot.json${RESET}"
pause "Apply migrations"

# ─── Step 4: Run migrations ─────────────────────────────────────────────────

banner "Run Migrations"
info "Apply all pending versioned migrations (V001-V005) and the"
info "repeatable migration (R__create_views.sql)."
echo ""
run drift migrate
pause "Inspect migration status"

# ─── Step 5: Inspect status ─────────────────────────────────────────────────

banner "Inspect Migration Status"
info "'drift info' shows which migrations have been applied,"
info "their versions, checksums, and timestamps."
echo ""
run drift info
pause "Validate migrations"

# ─── Step 6: Validate ───────────────────────────────────────────────────────

banner "Validate"
info "Check that applied migrations match the local files."
info "This catches modified or missing migration files."
echo ""
run drift validate
pause "Detect schema drift with diff"

# ─── Step 7: Schema diff ────────────────────────────────────────────────────

banner "Schema Diff"
info "Compare the live database against the pre-migration snapshot."
info "This shows exactly what the migrations changed — new tables, columns,"
info "and indexes that were added."
echo ""
run drift diff snapshot.json
echo ""
info "You can also generate the full schema as SQL:"
echo ""
run drift diff --format sql
pause "Lint migration files"

# ─── Step 8: Lint ────────────────────────────────────────────────────────────

banner "Lint Migration Files"
info "Check migration files for common issues like DROP TABLE,"
info "naming conventions, and other dangerous patterns."
echo ""
run drift lint
pause "Roll back a migration (no undo files needed)"

# ─── Step 9: Rollback ───────────────────────────────────────────────────────

banner "Rollback (no undo files needed)"
info "'drift rollback' reverses migrations by auto-generating reverse DDL."
info "Unlike 'drift undo', it does NOT require U__ undo files."
info ""
info "Let's roll back the last migration (V005 — add_indexes)."
echo ""
run drift rollback
echo ""
info "Check the status — V005 should now be Pending:"
echo ""
run drift info
pause "Roll back multiple migrations at once"

# ─── Step 10: Rollback multiple ─────────────────────────────────────────────

banner "Rollback Multiple"
info "Roll back 2 more migrations at once."
echo ""
run drift rollback --count 2
echo ""
run drift info
pause "Re-apply all migrations"

# ─── Step 11: Re-apply ──────────────────────────────────────────────────────

banner "Re-apply Migrations"
info "Run migrate again to bring everything back up."
echo ""
run drift migrate
echo ""
run drift info
pause "Preview a rollback with dry-run"

# ─── Step 12: Dry Run ───────────────────────────────────────────────────────

banner "Dry Run"
info "Preview what rollback would do, without executing anything."
echo ""
run drift rollback --dry-run --count 2
pause "Clean up and finish"

# ─── Step 13: Teardown ───────────────────────────────────────────────────────

banner "Teardown"
info "All done! Stopping Postgres and cleaning up."
echo ""
run docker compose down -v
rm -f drift.yaml snapshot.json migrations/U*.sql

echo ""
echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════════╗${RESET}"
echo -e "${BOLD}${CYAN}║           Tutorial complete!                         ║${RESET}"
echo -e "${BOLD}${CYAN}║                                                      ║${RESET}"
echo -e "${BOLD}${CYAN}║  You've used: migrate, info, validate, snapshot,     ║${RESET}"
echo -e "${BOLD}${CYAN}║  diff, rollback, dry-run, and lint.                  ║${RESET}"
echo -e "${BOLD}${CYAN}║                                                      ║${RESET}"
echo -e "${BOLD}${CYAN}║  Run 'drift tutorial' again anytime to repeat.       ║${RESET}"
echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════════╝${RESET}"
echo ""
