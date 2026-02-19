package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankchan/drift/internal/config"
	"github.com/frankchan/drift/internal/database"
	_ "github.com/frankchan/drift/internal/database/sqlite"
)

func setupTestEngine(t *testing.T) (*Engine, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migDir, 0755)

	cfg := &config.Config{
		URL:       dbPath,
		Locations: []string{migDir},
		Table:     "drift_schema_history",
	}

	ctx := context.Background()
	db, err := database.Open(ctx, "sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}

	eng := New(cfg, db)
	eng.SetOutput(os.Stderr)

	cleanup := func() {
		db.Close()
	}

	return eng, cleanup
}

func writeMigration(t *testing.T, dir, filename, sql string) {
	t.Helper()
	os.WriteFile(filepath.Join(dir, filename), []byte(sql), 0644)
}

func TestMigrate(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	migDir := eng.cfg.Locations[0]
	writeMigration(t, migDir, "V001__create_users.sql", `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL
		);
	`)
	writeMigration(t, migDir, "V002__create_posts.sql", `
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			body TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)

	ctx := context.Background()
	results, err := eng.Migrate(ctx, PlanOptions{})
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("migration %s failed: %v", r.Migration.Script, r.Error)
		}
	}

	// Running again should be a no-op
	results, err = eng.Migrate(ctx, PlanOptions{})
	if err != nil {
		t.Fatalf("second Migrate failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results on second run, got %d", len(results))
	}
}

func TestMigrateAndInfo(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	migDir := eng.cfg.Locations[0]
	writeMigration(t, migDir, "V001__first.sql", "CREATE TABLE t1 (id INTEGER);")
	writeMigration(t, migDir, "V002__second.sql", "CREATE TABLE t2 (id INTEGER);")

	ctx := context.Background()
	eng.Migrate(ctx, PlanOptions{})

	infos, err := eng.Info(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 2 {
		t.Fatalf("expected 2 infos, got %d", len(infos))
	}
	for _, info := range infos {
		if info.State != "Applied" {
			t.Errorf("expected Applied, got %s for V%s", info.State, info.Version)
		}
	}
}

func TestValidate(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	migDir := eng.cfg.Locations[0]
	writeMigration(t, migDir, "V001__init.sql", "CREATE TABLE t1 (id INTEGER);")

	ctx := context.Background()
	eng.Migrate(ctx, PlanOptions{})

	// No errors
	errors, err := eng.Validate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(errors) != 0 {
		t.Fatalf("expected 0 validation errors, got %d", len(errors))
	}

	// Modify the file to create a checksum mismatch
	writeMigration(t, migDir, "V001__init.sql", "CREATE TABLE t1 (id INTEGER, name TEXT);")

	errors, err = eng.Validate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(errors) != 1 {
		t.Fatalf("expected 1 validation error, got %d", len(errors))
	}
}

func TestRepair(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	migDir := eng.cfg.Locations[0]
	writeMigration(t, migDir, "V001__init.sql", "CREATE TABLE t1 (id INTEGER);")

	ctx := context.Background()
	eng.Migrate(ctx, PlanOptions{})

	// Modify file
	writeMigration(t, migDir, "V001__init.sql", "CREATE TABLE t1 (id INTEGER, name TEXT);")

	// Validate should fail
	errors, _ := eng.Validate(ctx)
	if len(errors) == 0 {
		t.Fatal("expected validation errors after modification")
	}

	// Repair
	if err := eng.Repair(ctx); err != nil {
		t.Fatal(err)
	}

	// Validate should pass now
	errors, _ = eng.Validate(ctx)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors after repair, got %d", len(errors))
	}
}

func TestBaseline(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	migDir := eng.cfg.Locations[0]
	writeMigration(t, migDir, "V001__old.sql", "CREATE TABLE old (id INTEGER);")
	writeMigration(t, migDir, "V002__new.sql", "CREATE TABLE new_table (id INTEGER);")

	ctx := context.Background()

	// Baseline at V001
	if err := eng.Baseline(ctx, "001"); err != nil {
		t.Fatal(err)
	}

	// Migrate should only apply V002
	results, err := eng.Migrate(ctx, PlanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 migration, got %d", len(results))
	}
	if results[0].Migration.Version != "002" {
		t.Errorf("expected V002, got V%s", results[0].Migration.Version)
	}
}

func TestDryRun(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	migDir := eng.cfg.Locations[0]
	writeMigration(t, migDir, "V001__init.sql", "CREATE TABLE t1 (id INTEGER);")

	ctx := context.Background()
	results, err := eng.Migrate(ctx, PlanOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 dry-run result, got %d", len(results))
	}

	// Table should NOT exist after dry run
	_, err = eng.db.DB().ExecContext(ctx, "SELECT 1 FROM t1")
	if err == nil {
		t.Error("table should not exist after dry run")
	}
}

func TestClean(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	migDir := eng.cfg.Locations[0]
	writeMigration(t, migDir, "V001__init.sql", "CREATE TABLE t1 (id INTEGER);")

	ctx := context.Background()
	eng.Migrate(ctx, PlanOptions{})

	if err := eng.Clean(ctx); err != nil {
		t.Fatal(err)
	}

	// History table should be gone
	_, err := eng.db.DB().ExecContext(ctx, "SELECT 1 FROM drift_schema_history")
	if err == nil {
		t.Error("history table should not exist after clean")
	}
}

func TestCherryPick(t *testing.T) {
	eng, cleanup := setupTestEngine(t)
	defer cleanup()

	migDir := eng.cfg.Locations[0]
	writeMigration(t, migDir, "V001__first.sql", "CREATE TABLE t1 (id INTEGER);")
	writeMigration(t, migDir, "V002__second.sql", "CREATE TABLE t2 (id INTEGER);")
	writeMigration(t, migDir, "V003__third.sql", "CREATE TABLE t3 (id INTEGER);")

	ctx := context.Background()
	results, err := eng.Migrate(ctx, PlanOptions{
		CherryPick: []string{"001", "003"},
		OutOfOrder: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 cherry-picked results, got %d", len(results))
	}
}

func TestSplitSQL(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"CREATE TABLE t (id INT); INSERT INTO t VALUES (1);", 2},
		{"-- comment\nCREATE TABLE t (id INT);", 1},
		{"SELECT 'hello;world'; SELECT 1;", 2},
		{"/* block ; comment */ SELECT 1;", 1},
		{"", 0},
	}

	for _, tt := range tests {
		got := splitSQL(tt.input)
		if len(got) != tt.want {
			t.Errorf("splitSQL(%q) = %d statements, want %d: %v", tt.input, len(got), tt.want, got)
		}
	}
}
