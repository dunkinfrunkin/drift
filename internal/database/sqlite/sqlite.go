package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/frankchan/drift/internal/database"
)

func init() {
	database.Register("sqlite", func(dsn string) database.Database {
		dsn = strings.TrimPrefix(dsn, "sqlite://")
		return &SQLite{dsn: dsn}
	})
}

// SQLite implements database.Database for SQLite.
type SQLite struct {
	dsn  string
	db   *sql.DB
	mu   sync.Mutex
	lock *os.File
}

func (s *SQLite) Open(ctx context.Context) error {
	// Ensure parent directory exists for non-memory databases
	if s.dsn != ":memory:" && s.dsn != "" {
		dir := filepath.Dir(s.dsn)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", s.dsn)
	if err != nil {
		return err
	}
	// Enable WAL mode and foreign keys
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return err
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return err
	}
	s.db = db
	return nil
}

func (s *SQLite) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLite) DB() *sql.DB { return s.db }

func (s *SQLite) Dialect() database.Dialect { return &Dialect{} }

// Lock uses an in-process mutex for SQLite (single-process assumed).
func (s *SQLite) Lock(_ context.Context) error {
	s.mu.Lock()
	return nil
}

func (s *SQLite) Unlock(_ context.Context) error {
	s.mu.Unlock()
	return nil
}

// Dialect implements database.Dialect for SQLite.
type Dialect struct{}

func (d *Dialect) Name() string { return "sqlite" }

func (d *Dialect) QuoteIdentifier(name string) string {
	return `"` + name + `"`
}

func (d *Dialect) Placeholder(_ int) string { return "?" }

func (d *Dialect) SupportsTransactionalDDL() bool { return true }

func (d *Dialect) CreateHistoryTableSQL(table string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    installed_rank INTEGER NOT NULL PRIMARY KEY,
    version        TEXT,
    description    TEXT NOT NULL,
    type           TEXT NOT NULL,
    script         TEXT NOT NULL,
    checksum       INTEGER,
    installed_by   TEXT NOT NULL,
    installed_on   TEXT NOT NULL DEFAULT (datetime('now')),
    execution_time INTEGER NOT NULL,
    success        INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS %s_s_idx ON %s (success);`,
		d.QuoteIdentifier(table),
		table, d.QuoteIdentifier(table))
}
