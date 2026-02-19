package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

// Database is the main interface for interacting with a target database.
type Database interface {
	// Open establishes a connection to the database.
	Open(ctx context.Context) error
	// Close closes the database connection.
	Close() error
	// DB returns the underlying *sql.DB.
	DB() *sql.DB
	// Dialect returns the SQL dialect for this database.
	Dialect() Dialect
	// Lock acquires an advisory lock to prevent concurrent migrations.
	Lock(ctx context.Context) error
	// Unlock releases the advisory lock.
	Unlock(ctx context.Context) error
}

// Dialect defines SQL dialect differences between databases.
type Dialect interface {
	// Name returns the driver name (postgres, mysql, sqlite).
	Name() string
	// QuoteIdentifier quotes a table or column name.
	QuoteIdentifier(name string) string
	// Placeholder returns the parameter placeholder for the given position (1-indexed).
	Placeholder(position int) string
	// SupportsTransactionalDDL returns true if DDL statements can be rolled back.
	SupportsTransactionalDDL() bool
	// CreateHistoryTableSQL returns the SQL to create the schema history table.
	CreateHistoryTableSQL(table string) string
}

// Factory creates a Database from a DSN.
type Factory func(dsn string) Database

var (
	mu        sync.RWMutex
	factories = make(map[string]Factory)
)

// Register adds a database driver factory.
func Register(name string, factory Factory) {
	mu.Lock()
	defer mu.Unlock()
	factories[name] = factory
}

// Open creates and opens a database connection for the given driver name and DSN.
func Open(ctx context.Context, driver, dsn string) (Database, error) {
	mu.RLock()
	factory, ok := factories[driver]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}

	db := factory(dsn)
	if err := db.Open(ctx); err != nil {
		return nil, fmt.Errorf("opening %s database: %w", driver, err)
	}
	return db, nil
}

// Drivers returns the names of all registered drivers.
func Drivers() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	return names
}
