package sqlite

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenAndClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s := &SQLite{dsn: dbPath}

	ctx := context.Background()
	if err := s.Open(ctx); err != nil {
		t.Fatal(err)
	}

	// Verify we can execute queries
	_, err := s.DB().ExecContext(ctx, "CREATE TABLE test (id INTEGER)")
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestDialect(t *testing.T) {
	d := &Dialect{}

	if d.Name() != "sqlite" {
		t.Errorf("Name() = %q", d.Name())
	}
	if d.QuoteIdentifier("foo") != `"foo"` {
		t.Errorf("QuoteIdentifier = %q", d.QuoteIdentifier("foo"))
	}
	if d.Placeholder(1) != "?" {
		t.Errorf("Placeholder = %q", d.Placeholder(1))
	}
	if !d.SupportsTransactionalDDL() {
		t.Error("SQLite should support transactional DDL")
	}

	sql := d.CreateHistoryTableSQL("test_history")
	if sql == "" {
		t.Error("empty CreateHistoryTableSQL")
	}
}

func TestLockUnlock(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s := &SQLite{dsn: dbPath}
	ctx := context.Background()
	s.Open(ctx)
	defer s.Close()

	if err := s.Lock(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.Unlock(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestMemoryDB(t *testing.T) {
	s := &SQLite{dsn: ":memory:"}
	ctx := context.Background()
	if err := s.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err := s.DB().ExecContext(ctx, "CREATE TABLE test (id INTEGER)")
	if err != nil {
		t.Fatal(err)
	}
}
