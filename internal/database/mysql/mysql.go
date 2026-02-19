package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/frankchan/drift/internal/database"
)

func init() {
	database.Register("mysql", func(dsn string) database.Database {
		// Strip mysql:// prefix if present; go-sql-driver expects user:pass@tcp(host)/db
		dsn = strings.TrimPrefix(dsn, "mysql://")
		return &MySQL{dsn: dsn}
	})
}

// MySQL implements database.Database for MySQL.
type MySQL struct {
	dsn string
	db  *sql.DB
}

func (m *MySQL) Open(ctx context.Context) error {
	db, err := sql.Open("mysql", m.dsn)
	if err != nil {
		return err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return err
	}
	m.db = db
	return nil
}

func (m *MySQL) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *MySQL) DB() *sql.DB { return m.db }

func (m *MySQL) Dialect() database.Dialect { return &Dialect{} }

func (m *MySQL) Lock(ctx context.Context) error {
	var result int
	err := m.db.QueryRowContext(ctx, "SELECT GET_LOCK('drift_migration', 10)").Scan(&result)
	if err != nil {
		return err
	}
	if result != 1 {
		return fmt.Errorf("could not acquire migration lock")
	}
	return nil
}

func (m *MySQL) Unlock(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, "SELECT RELEASE_LOCK('drift_migration')")
	return err
}

// Dialect implements database.Dialect for MySQL.
type Dialect struct{}

func (d *Dialect) Name() string { return "mysql" }

func (d *Dialect) QuoteIdentifier(name string) string {
	return "`" + name + "`"
}

func (d *Dialect) Placeholder(_ int) string { return "?" }

func (d *Dialect) SupportsTransactionalDDL() bool { return false }

func (d *Dialect) CreateHistoryTableSQL(table string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    installed_rank INTEGER NOT NULL,
    version        VARCHAR(50),
    description    VARCHAR(200) NOT NULL,
    type           VARCHAR(20) NOT NULL,
    script         VARCHAR(1000) NOT NULL,
    checksum       INTEGER,
    installed_by   VARCHAR(100) NOT NULL,
    installed_on   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    execution_time INTEGER NOT NULL,
    success        BOOLEAN NOT NULL,
    PRIMARY KEY (installed_rank),
    INDEX %s_s_idx (success)
) ENGINE=InnoDB;`, d.QuoteIdentifier(table), table)
}
