package diff

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/frankchan/drift/internal/database"
	_ "github.com/frankchan/drift/internal/database/sqlite"
)

func TestCaptureSQLiteSchema(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(ctx, "sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create some tables
	db.DB().ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT)")
	db.DB().ExecContext(ctx, "CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER REFERENCES users(id), title TEXT NOT NULL)")
	db.DB().ExecContext(ctx, "CREATE INDEX idx_posts_user ON posts(user_id)")

	snap, err := CaptureSchema(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	if len(snap.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(snap.Tables))
	}

	// Find users table
	var usersTable *Table
	for i := range snap.Tables {
		if snap.Tables[i].Name == "users" {
			usersTable = &snap.Tables[i]
			break
		}
	}
	if usersTable == nil {
		t.Fatal("users table not found")
	}
	if len(usersTable.Columns) != 3 {
		t.Errorf("expected 3 columns in users, got %d", len(usersTable.Columns))
	}

	// Check indexes
	if len(snap.Indexes) < 1 {
		t.Error("expected at least 1 index")
	}
}

func TestLoadSnapshot(t *testing.T) {
	snap := &SchemaSnapshot{
		Tables: []Table{
			{Name: "t1", Columns: []Column{{Name: "id", DataType: "INTEGER"}}},
		},
	}

	path := filepath.Join(t.TempDir(), "snap.json")
	data, _ := json.Marshal(snap)
	os.WriteFile(path, data, 0644)

	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Tables) != 1 || loaded.Tables[0].Name != "t1" {
		t.Errorf("unexpected loaded snapshot: %+v", loaded)
	}
}
