package history

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/frankchan/drift/internal/database"
	_ "github.com/frankchan/drift/internal/database/sqlite"
	"github.com/frankchan/drift/internal/migration"
)

func openTestDB(t *testing.T) (database.Database, func()) {
	t.Helper()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(ctx, "sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	return db, func() { db.Close() }
}

func TestEnsureTableAndRecord(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	ctx := context.Background()

	h := New(db, "drift_schema_history")
	if err := h.EnsureTable(ctx); err != nil {
		t.Fatal(err)
	}

	m := &migration.Migration{
		Version:     "001",
		Description: "create users",
		Type:        migration.TypeVersioned,
		Script:      "V001__create_users.sql",
		Checksum:    12345,
	}

	if err := h.Record(ctx, m, 50*time.Millisecond, true); err != nil {
		t.Fatal(err)
	}

	all, err := h.All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 record, got %d", len(all))
	}
	if all[0].Version != "001" {
		t.Errorf("version = %q, want %q", all[0].Version, "001")
	}
	if all[0].Checksum != 12345 {
		t.Errorf("checksum = %d, want %d", all[0].Checksum, 12345)
	}
	if !all[0].Success {
		t.Error("expected success = true")
	}
}

func TestRecordMultipleAndOrdering(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	ctx := context.Background()

	h := New(db, "drift_schema_history")
	h.EnsureTable(ctx)

	for i, v := range []string{"001", "002", "003"} {
		m := &migration.Migration{
			Version:     v,
			Description: "migration " + v,
			Type:        migration.TypeVersioned,
			Script:      "V" + v + "__m.sql",
			Checksum:    uint32(i),
		}
		h.Record(ctx, m, time.Millisecond, true)
	}

	all, err := h.All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	// Should be ordered by installed_rank
	if all[0].InstalledRank != 1 || all[1].InstalledRank != 2 || all[2].InstalledRank != 3 {
		t.Errorf("unexpected rank ordering: %d, %d, %d", all[0].InstalledRank, all[1].InstalledRank, all[2].InstalledRank)
	}
}

func TestRemove(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	ctx := context.Background()

	h := New(db, "drift_schema_history")
	h.EnsureTable(ctx)

	m := &migration.Migration{Version: "001", Description: "test", Type: migration.TypeVersioned, Script: "V001__test.sql"}
	h.Record(ctx, m, 0, true)

	if err := h.Remove(ctx, "001"); err != nil {
		t.Fatal(err)
	}

	all, _ := h.All(ctx)
	if len(all) != 0 {
		t.Fatalf("expected 0 after remove, got %d", len(all))
	}
}

func TestUpdateChecksum(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	ctx := context.Background()

	h := New(db, "drift_schema_history")
	h.EnsureTable(ctx)

	m := &migration.Migration{Version: "001", Description: "test", Type: migration.TypeVersioned, Script: "V001__test.sql", Checksum: 111}
	h.Record(ctx, m, 0, true)

	h.UpdateChecksum(ctx, "001", 999)

	all, _ := h.All(ctx)
	if all[0].Checksum != 999 {
		t.Errorf("checksum = %d, want 999", all[0].Checksum)
	}
}

func TestClearAndDrop(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	ctx := context.Background()

	h := New(db, "drift_schema_history")
	h.EnsureTable(ctx)

	m := &migration.Migration{Version: "001", Description: "test", Type: migration.TypeVersioned, Script: "V001__test.sql"}
	h.Record(ctx, m, 0, true)

	h.Clear(ctx)
	all, _ := h.All(ctx)
	if len(all) != 0 {
		t.Fatalf("expected 0 after clear, got %d", len(all))
	}

	// Table should still exist after clear
	h.Record(ctx, m, 0, true)
	all, _ = h.All(ctx)
	if len(all) != 1 {
		t.Fatal("table should still exist after clear")
	}

	// Drop should remove the table
	h.Drop(ctx)
	_, err := h.All(ctx)
	if err == nil {
		t.Error("expected error after dropping table")
	}
}

func TestMarkSuccess(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()
	ctx := context.Background()

	h := New(db, "drift_schema_history")
	h.EnsureTable(ctx)

	m := &migration.Migration{Version: "001", Description: "test", Type: migration.TypeVersioned, Script: "V001__test.sql"}
	h.Record(ctx, m, 0, false) // record as failed

	all, _ := h.All(ctx)
	if all[0].Success {
		t.Error("expected success=false")
	}

	h.MarkSuccess(ctx, "001", true)
	all, _ = h.All(ctx)
	if !all[0].Success {
		t.Error("expected success=true after mark")
	}
}
