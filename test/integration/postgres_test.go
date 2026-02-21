//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVersion(t *testing.T) {
	passed := false
	defer func() { recordResult("version", passed, false) }()

	stdout, _, err := runCLI(t, "version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(strings.ToLower(stdout), "drift") {
		t.Fatalf("expected 'drift' in output, got: %s", stdout)
	}
	passed = true
}

func TestInit(t *testing.T) {
	passed := false
	defer func() { recordResult("init", passed, false) }()

	t.Run("creates_config_and_dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		stdout, _, err := runCLI(t, "init", "--driver", "postgres")
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}
		if !strings.Contains(stdout, "Initialized") {
			t.Fatalf("expected 'Initialized' in output, got: %s", stdout)
		}

		// Check drift.yaml exists
		if _, err := os.Stat(filepath.Join(tmpDir, "drift.yaml")); err != nil {
			t.Fatalf("drift.yaml not created: %v", err)
		}
		// Check migrations dir exists
		if _, err := os.Stat(filepath.Join(tmpDir, "migrations")); err != nil {
			t.Fatalf("migrations/ not created: %v", err)
		}
	})

	t.Run("fails_if_exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(origDir)

		// Create drift.yaml first
		os.WriteFile(filepath.Join(tmpDir, "drift.yaml"), []byte("url: test"), 0644)

		_, _, err := runCLI(t, "init")
		if err == nil {
			t.Fatal("expected error when drift.yaml already exists")
		}
	})

	passed = true
}

func TestMigrate(t *testing.T) {
	passed := false
	defer func() { recordResult("migrate", passed, false) }()

	t.Run("applies_pending", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)

		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		writeMigration(t, migDir, "V002__create_posts.sql",
			"CREATE TABLE posts (id SERIAL PRIMARY KEY, user_id INT REFERENCES users(id), title TEXT);")

		stdout, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("migrate failed: %v\nstdout: %s", err, stdout)
		}
		if !strings.Contains(stdout, "Successfully applied 2 migration") {
			t.Fatalf("expected 2 migrations applied, got: %s", stdout)
		}
	})

	t.Run("idempotent_rerun", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)

		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")

		// First run
		_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("first migrate failed: %v", err)
		}

		// Second run - should be no-op
		stdout, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("second migrate failed: %v", err)
		}
		if !strings.Contains(stdout, "up to date") {
			t.Fatalf("expected 'up to date', got: %s", stdout)
		}
	})

	t.Run("dry_run", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)

		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")

		stdout, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir, "--dry-run")
		if err != nil {
			t.Fatalf("dry-run failed: %v\nstdout: %s", err, stdout)
		}

		// After dry-run, running again should still find pending migrations
		stdout2, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("migrate after dry-run failed: %v", err)
		}
		if !strings.Contains(stdout2, "Successfully applied 1 migration") {
			t.Fatalf("expected 1 migration still pending after dry-run, got: %s", stdout2)
		}
	})

	t.Run("target_version", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)

		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		writeMigration(t, migDir, "V002__create_posts.sql",
			"CREATE TABLE posts (id SERIAL PRIMARY KEY, title TEXT);")
		writeMigration(t, migDir, "V003__create_tags.sql",
			"CREATE TABLE tags (id SERIAL PRIMARY KEY, name TEXT);")

		stdout, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir, "--target", "002")
		if err != nil {
			t.Fatalf("migrate with target failed: %v\nstdout: %s", err, stdout)
		}
		if !strings.Contains(stdout, "Successfully applied 2 migration") {
			t.Fatalf("expected 2 migrations, got: %s", stdout)
		}
	})

	t.Run("cherry_pick_out_of_order", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)

		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		writeMigration(t, migDir, "V002__create_posts.sql",
			"CREATE TABLE posts (id SERIAL PRIMARY KEY, title TEXT);")
		writeMigration(t, migDir, "V003__create_tags.sql",
			"CREATE TABLE tags (id SERIAL PRIMARY KEY, name TEXT);")

		// Cherry-pick only V003, skipping V001 and V002
		stdout, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir,
			"--cherry-pick", "003", "--out-of-order")
		if err != nil {
			t.Fatalf("cherry-pick failed: %v\nstdout: %s", err, stdout)
		}
		if !strings.Contains(stdout, "Successfully applied 1 migration") {
			t.Fatalf("expected 1 migration, got: %s", stdout)
		}
	})

	t.Run("skip", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)

		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		writeMigration(t, migDir, "V002__create_posts.sql",
			"CREATE TABLE posts (id SERIAL PRIMARY KEY, title TEXT);")
		writeMigration(t, migDir, "V003__create_tags.sql",
			"CREATE TABLE tags (id SERIAL PRIMARY KEY, name TEXT);")

		stdout, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir, "--skip", "002")
		if err != nil {
			t.Fatalf("migrate with skip failed: %v\nstdout: %s", err, stdout)
		}
		if !strings.Contains(stdout, "Successfully applied 2 migration") {
			t.Fatalf("expected 2 migrations (001,003), got: %s", stdout)
		}
	})

	passed = true
}

func TestUndo(t *testing.T) {
	passed := false
	defer func() { recordResult("undo", passed, false) }()

	t.Run("undo_last", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)

		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		writeMigration(t, migDir, "U001__drop_users.sql",
			"DROP TABLE IF EXISTS users;")

		// Apply first
		_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("migrate failed: %v", err)
		}

		// Undo
		stdout, _, err := runCLI(t, "undo", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("undo failed: %v\nstdout: %s", err, stdout)
		}
		if !strings.Contains(stdout, "Successfully undone 1 migration") {
			t.Fatalf("expected 1 undone, got: %s", stdout)
		}
	})

	t.Run("undo_count_2", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)

		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		writeMigration(t, migDir, "V002__create_posts.sql",
			"CREATE TABLE posts (id SERIAL PRIMARY KEY, title TEXT);")
		writeMigration(t, migDir, "U001__drop_users.sql",
			"DROP TABLE IF EXISTS users;")
		writeMigration(t, migDir, "U002__drop_posts.sql",
			"DROP TABLE IF EXISTS posts;")

		_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("migrate failed: %v", err)
		}

		stdout, _, err := runCLI(t, "undo", "--url", pgURL, "--locations", migDir, "--count", "2")
		if err != nil {
			t.Fatalf("undo count 2 failed: %v\nstdout: %s", err, stdout)
		}
		if !strings.Contains(stdout, "Successfully undone 2 migration") {
			t.Fatalf("expected 2 undone, got: %s", stdout)
		}
	})

	t.Run("dry_run_preserves_state", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)

		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		writeMigration(t, migDir, "U001__drop_users.sql",
			"DROP TABLE IF EXISTS users;")

		_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("migrate failed: %v", err)
		}

		// Dry-run undo
		_, _, err = runCLI(t, "undo", "--url", pgURL, "--locations", migDir, "--dry-run")
		if err != nil {
			t.Fatalf("undo dry-run failed: %v", err)
		}

		// Info should still show Applied
		stdout, _, err := runCLI(t, "info", "--url", pgURL, "--locations", migDir, "--format", "json")
		if err != nil {
			t.Fatalf("info failed: %v", err)
		}
		if !strings.Contains(stdout, "Applied") {
			t.Fatalf("expected Applied state after dry-run undo, got: %s", stdout)
		}
	})

	passed = true
}

func TestInfo(t *testing.T) {
	passed := false
	defer func() { recordResult("info", passed, false) }()

	pgURL, migDir := setupTestSchema(t)
	writeMigration(t, migDir, "V001__create_users.sql",
		"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
	_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	t.Run("format_table", func(t *testing.T) {
		stdout, _, err := runCLI(t, "info", "--url", pgURL, "--locations", migDir, "--format", "table")
		if err != nil {
			t.Fatalf("info table failed: %v", err)
		}
		if !strings.Contains(stdout, "001") || !strings.Contains(stdout, "create users") {
			t.Fatalf("table output missing expected content: %s", stdout)
		}
	})

	t.Run("format_json", func(t *testing.T) {
		stdout, _, err := runCLI(t, "info", "--url", pgURL, "--locations", migDir, "--format", "json")
		if err != nil {
			t.Fatalf("info json failed: %v", err)
		}
		var infos []map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &infos); err != nil {
			t.Fatalf("invalid json: %v\noutput: %s", err, stdout)
		}
		if len(infos) == 0 {
			t.Fatal("expected at least 1 migration info")
		}
	})

	t.Run("format_yaml", func(t *testing.T) {
		stdout, _, err := runCLI(t, "info", "--url", pgURL, "--locations", migDir, "--format", "yaml")
		if err != nil {
			t.Fatalf("info yaml failed: %v", err)
		}
		if !strings.Contains(stdout, "version:") || !strings.Contains(stdout, "state:") {
			t.Fatalf("yaml output missing expected fields: %s", stdout)
		}
	})

	passed = true
}

func TestValidate(t *testing.T) {
	passed := false
	defer func() { recordResult("validate", passed, false) }()

	t.Run("passes_clean", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)
		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")

		_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("migrate failed: %v", err)
		}

		stdout, _, err := runCLI(t, "validate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("validate failed: %v\nstdout: %s", err, stdout)
		}
	})

	t.Run("fails_on_checksum_mismatch", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)
		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")

		_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("migrate failed: %v", err)
		}

		// Modify the file to cause checksum mismatch
		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL, email TEXT);")

		_, _, err = runCLI(t, "validate", "--url", pgURL, "--locations", migDir)
		if err == nil {
			t.Fatal("expected validation to fail on checksum mismatch")
		}
	})

	passed = true
}

func TestClean(t *testing.T) {
	passed := false
	defer func() { recordResult("clean", passed, false) }()

	pgURL, migDir := setupTestSchema(t)
	writeMigration(t, migDir, "V001__create_users.sql",
		"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")

	_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Clean with --yes to skip confirmation
	stdout, _, err := runCLI(t, "clean", "--url", pgURL, "--locations", migDir, "--yes")
	if err != nil {
		t.Fatalf("clean failed: %v\nstdout: %s", err, stdout)
	}
	if !strings.Contains(stdout, "dropped") {
		t.Fatalf("expected 'dropped' in output, got: %s", stdout)
	}

	passed = true
}

func TestBaseline(t *testing.T) {
	passed := false
	defer func() { recordResult("baseline", passed, false) }()

	t.Run("version_skips_earlier", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)
		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
		writeMigration(t, migDir, "V002__create_posts.sql",
			"CREATE TABLE posts (id SERIAL PRIMARY KEY, title TEXT);")

		// Baseline at version 001
		stdout, _, err := runCLI(t, "baseline", "--url", pgURL, "--locations", migDir, "--version", "001")
		if err != nil {
			t.Fatalf("baseline failed: %v\nstdout: %s", err, stdout)
		}
		if !strings.Contains(stdout, "Baselined") {
			t.Fatalf("expected 'Baselined' in output, got: %s", stdout)
		}

		// Migrate should only apply V002
		stdout, _, err = runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("migrate after baseline failed: %v\nstdout: %s", err, stdout)
		}
		if !strings.Contains(stdout, "Successfully applied 1 migration") {
			t.Fatalf("expected 1 migration after baseline, got: %s", stdout)
		}
	})

	t.Run("from_db_output", func(t *testing.T) {
		pgURL, migDir := setupTestSchema(t)
		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")

		// Apply migrations first so there's a schema to capture
		_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
		if err != nil {
			t.Fatalf("migrate failed: %v", err)
		}

		outFile := filepath.Join(t.TempDir(), "baseline.sql")
		stdout, _, err := runCLI(t, "baseline", "--url", pgURL, "--locations", migDir,
			"--from-db", "--output", outFile, "--version", "001")
		if err != nil {
			t.Fatalf("baseline --from-db failed: %v\nstdout: %s", err, stdout)
		}

		data, err := os.ReadFile(outFile)
		if err != nil {
			t.Fatalf("read output file: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("baseline output file is empty")
		}
	})

	passed = true
}

func TestRepair(t *testing.T) {
	passed := false
	defer func() { recordResult("repair", passed, false) }()

	pgURL, migDir := setupTestSchema(t)
	writeMigration(t, migDir, "V001__create_users.sql",
		"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")

	// Apply migration
	_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Modify migration to create checksum mismatch
	writeMigration(t, migDir, "V001__create_users.sql",
		"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL, email TEXT);")

	// Validate should fail
	_, _, err = runCLI(t, "validate", "--url", pgURL, "--locations", migDir)
	if err == nil {
		t.Fatal("expected validate to fail before repair")
	}

	// Repair
	stdout, _, err := runCLI(t, "repair", "--url", pgURL, "--locations", migDir)
	if err != nil {
		t.Fatalf("repair failed: %v\nstdout: %s", err, stdout)
	}
	if !strings.Contains(stdout, "Repaired") {
		t.Fatalf("expected 'Repaired' in output, got: %s", stdout)
	}

	// Validate should pass now
	_, _, err = runCLI(t, "validate", "--url", pgURL, "--locations", migDir)
	if err != nil {
		t.Fatalf("validate after repair failed: %v", err)
	}

	passed = true
}

func TestDiff(t *testing.T) {
	passed := false
	defer func() { recordResult("diff", passed, false) }()

	pgURL, migDir := setupTestSchema(t)
	writeMigration(t, migDir, "V001__create_users.sql",
		"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
	_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	t.Run("text_format", func(t *testing.T) {
		stdout, _, err := runCLI(t, "diff", "--url", pgURL, "--locations", migDir, "--format", "text")
		if err != nil {
			t.Fatalf("diff text failed: %v", err)
		}
		// Should show some diff against empty schema (tables were created)
		if len(stdout) == 0 {
			t.Fatal("expected non-empty diff output")
		}
	})

	t.Run("json_format", func(t *testing.T) {
		stdout, _, err := runCLI(t, "diff", "--url", pgURL, "--locations", migDir, "--format", "json")
		if err != nil {
			t.Fatalf("diff json failed: %v", err)
		}
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			t.Fatalf("invalid json: %v\noutput: %s", err, stdout)
		}
	})

	t.Run("sql_format", func(t *testing.T) {
		stdout, _, err := runCLI(t, "diff", "--url", pgURL, "--locations", migDir, "--format", "sql")
		if err != nil {
			t.Fatalf("diff sql failed: %v", err)
		}
		_ = stdout // SQL format may be empty if no actionable changes
	})

	t.Run("diff_against_snapshot", func(t *testing.T) {
		// First take a snapshot
		snapFile := filepath.Join(t.TempDir(), "snap.json")
		_, _, err := runCLI(t, "snapshot", "--url", pgURL, "--locations", migDir, "--output", snapFile)
		if err != nil {
			t.Fatalf("snapshot failed: %v", err)
		}

		// Diff against the snapshot (should show no differences)
		stdout, _, err := runCLI(t, "diff", "--url", pgURL, "--locations", migDir, snapFile)
		if err != nil {
			t.Fatalf("diff against snapshot failed: %v", err)
		}
		if !strings.Contains(stdout, "No differences") {
			t.Fatalf("expected no differences, got: %s", stdout)
		}
	})

	passed = true
}

func TestSnapshot(t *testing.T) {
	passed := false
	defer func() { recordResult("snapshot", passed, false) }()

	pgURL, migDir := setupTestSchema(t)
	writeMigration(t, migDir, "V001__create_users.sql",
		"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
	_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	t.Run("json_stdout", func(t *testing.T) {
		stdout, _, err := runCLI(t, "snapshot", "--url", pgURL, "--locations", migDir, "--format", "json")
		if err != nil {
			t.Fatalf("snapshot json failed: %v", err)
		}
		var snap map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &snap); err != nil {
			t.Fatalf("invalid json: %v\noutput: %s", err, stdout)
		}
	})

	t.Run("yaml_stdout", func(t *testing.T) {
		stdout, _, err := runCLI(t, "snapshot", "--url", pgURL, "--locations", migDir, "--format", "yaml")
		if err != nil {
			t.Fatalf("snapshot yaml failed: %v", err)
		}
		if len(stdout) == 0 {
			t.Fatal("expected non-empty yaml output")
		}
	})

	t.Run("output_to_file", func(t *testing.T) {
		outFile := filepath.Join(t.TempDir(), "snapshot.json")
		_, _, err := runCLI(t, "snapshot", "--url", pgURL, "--locations", migDir, "--output", outFile)
		if err != nil {
			t.Fatalf("snapshot to file failed: %v", err)
		}
		data, err := os.ReadFile(outFile)
		if err != nil {
			t.Fatalf("read snapshot file: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("snapshot file is empty")
		}
	})

	passed = true
}

func TestLint(t *testing.T) {
	passed := false
	defer func() { recordResult("lint", passed, false) }()

	t.Run("clean_passes", func(t *testing.T) {
		_, migDir := setupTestSchema(t)
		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")

		cfgDir := t.TempDir()
		cfgPath := writeConfig(t, cfgDir, "postgres://unused", migDir)

		stdout, _, err := runCLI(t, "lint", "--config", cfgPath)
		if err != nil {
			t.Fatalf("lint clean failed: %v\nstdout: %s", err, stdout)
		}
	})

	t.Run("detects_drop_table", func(t *testing.T) {
		_, migDir := setupTestSchema(t)
		writeMigration(t, migDir, "V001__drop_stuff.sql",
			"DROP TABLE users;")

		cfgDir := t.TempDir()
		cfgPath := writeConfig(t, cfgDir, "postgres://unused", migDir)

		stdout, _, err := runCLI(t, "lint", "--config", cfgPath)
		if err == nil {
			t.Fatalf("expected lint to fail on DROP TABLE, stdout: %s", stdout)
		}
		if !strings.Contains(stdout, "DROP TABLE") || !strings.Contains(stdout, "no-drop-table") {
			t.Fatalf("expected DROP TABLE warning, got: %s", stdout)
		}
	})

	t.Run("rules_filtering", func(t *testing.T) {
		_, migDir := setupTestSchema(t)
		// This has DROP TABLE (error) but we only enable naming-convention rule
		writeMigration(t, migDir, "V001__drop_stuff.sql",
			"DROP TABLE users;")

		cfgDir := t.TempDir()
		cfgPath := writeConfig(t, cfgDir, "postgres://unused", migDir)

		// Only check naming-convention - should pass since it has a description
		stdout, _, err := runCLI(t, "lint", "--config", cfgPath, "--rules", "naming-convention")
		if err != nil {
			t.Fatalf("lint with rules filter failed: %v\nstdout: %s", err, stdout)
		}
	})

	passed = true
}

func TestSquash(t *testing.T) {
	passed := false
	defer func() { recordResult("squash", passed, false) }()

	t.Run("all", func(t *testing.T) {
		_, migDir := setupTestSchema(t)
		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY);")
		writeMigration(t, migDir, "V002__create_posts.sql",
			"CREATE TABLE posts (id SERIAL PRIMARY KEY);")

		cfgDir := t.TempDir()
		cfgPath := writeConfig(t, cfgDir, "postgres://unused", migDir)

		stdout, _, err := runCLI(t, "squash", "--config", cfgPath)
		if err != nil {
			t.Fatalf("squash failed: %v\nstdout: %s", err, stdout)
		}
		if !strings.Contains(stdout, "V001__create_users.sql") || !strings.Contains(stdout, "V002__create_posts.sql") {
			t.Fatalf("expected both migration comments in squashed output, got: %s", stdout)
		}
	})

	t.Run("up_to", func(t *testing.T) {
		_, migDir := setupTestSchema(t)
		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY);")
		writeMigration(t, migDir, "V002__create_posts.sql",
			"CREATE TABLE posts (id SERIAL PRIMARY KEY);")
		writeMigration(t, migDir, "V003__create_tags.sql",
			"CREATE TABLE tags (id SERIAL PRIMARY KEY);")

		cfgDir := t.TempDir()
		cfgPath := writeConfig(t, cfgDir, "postgres://unused", migDir)

		stdout, _, err := runCLI(t, "squash", "--config", cfgPath, "--up-to", "002")
		if err != nil {
			t.Fatalf("squash up-to failed: %v", err)
		}
		if strings.Contains(stdout, "V003") {
			t.Fatalf("squash --up-to 002 should not include V003, got: %s", stdout)
		}
		if !strings.Contains(stdout, "V001") || !strings.Contains(stdout, "V002") {
			t.Fatalf("squash --up-to 002 should include V001 and V002, got: %s", stdout)
		}
	})

	t.Run("output_to_file", func(t *testing.T) {
		_, migDir := setupTestSchema(t)
		writeMigration(t, migDir, "V001__create_users.sql",
			"CREATE TABLE users (id SERIAL PRIMARY KEY);")

		cfgDir := t.TempDir()
		cfgPath := writeConfig(t, cfgDir, "postgres://unused", migDir)
		outFile := filepath.Join(t.TempDir(), "squashed.sql")

		stdout, _, err := runCLI(t, "squash", "--config", cfgPath, "--output", outFile)
		if err != nil {
			t.Fatalf("squash to file failed: %v\nstdout: %s", err, stdout)
		}
		data, err := os.ReadFile(outFile)
		if err != nil {
			t.Fatalf("read squash output: %v", err)
		}
		if len(data) == 0 {
			t.Fatal("squash output file is empty")
		}
	})

	passed = true
}

func TestReport(t *testing.T) {
	passed := false
	defer func() { recordResult("report", passed, false) }()

	pgURL, migDir := setupTestSchema(t)
	writeMigration(t, migDir, "V001__create_users.sql",
		"CREATE TABLE users (id SERIAL PRIMARY KEY, name TEXT NOT NULL);")
	_, _, err := runCLI(t, "migrate", "--url", pgURL, "--locations", migDir)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	t.Run("text_format", func(t *testing.T) {
		stdout, _, err := runCLI(t, "report", "--url", pgURL, "--locations", migDir, "--format", "text")
		if err != nil {
			t.Fatalf("report text failed: %v", err)
		}
		if !strings.Contains(stdout, "Migration Report") {
			t.Fatalf("expected 'Migration Report' in output, got: %s", stdout)
		}
	})

	t.Run("json_format", func(t *testing.T) {
		stdout, _, err := runCLI(t, "report", "--url", pgURL, "--locations", migDir, "--format", "json")
		if err != nil {
			t.Fatalf("report json failed: %v", err)
		}
		var report map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &report); err != nil {
			t.Fatalf("invalid json: %v\noutput: %s", err, stdout)
		}
		if _, ok := report["applied"]; !ok {
			t.Fatal("expected 'applied' field in report json")
		}
	})

	t.Run("html_format", func(t *testing.T) {
		stdout, _, err := runCLI(t, "report", "--url", pgURL, "--locations", migDir, "--format", "html")
		if err != nil {
			t.Fatalf("report html failed: %v", err)
		}
		if !strings.Contains(stdout, "<!DOCTYPE html>") {
			t.Fatalf("expected HTML output, got: %s", stdout)
		}
	})

	passed = true
}

func TestUI(t *testing.T) {
	passed := false
	defer func() { recordResult("ui", passed, false) }()

	pgURL, migDir := setupTestSchema(t)

	// Write a config file for the UI server
	cfgDir := t.TempDir()
	cfgPath := writeConfig(t, cfgDir, pgURL, migDir)

	port := "14077"

	// Start the UI server in a goroutine - it blocks until stopped
	errCh := make(chan error, 1)
	go func() {
		_, _, err := runCLI(t, "ui", "--config", cfgPath, "--port", port, "--no-open")
		errCh <- err
	}()

	// Wait for the server to start
	var lastErr error
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/api/v1/config", port))
		if err == nil {
			if resp.StatusCode == 200 {
				resp.Body.Close()
				passed = true
				return
			}
			resp.Body.Close()
		}
		lastErr = err
		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("UI server did not respond with 200 within 5s, last error: %v", lastErr)
}
