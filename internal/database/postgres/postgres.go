package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/frankchan/drift/internal/database"
)

func init() {
	database.Register("postgres", func(dsn string) database.Database {
		return &Postgres{dsn: dsn}
	})
}

// Postgres implements database.Database for PostgreSQL.
type Postgres struct {
	dsn string
	db  *sql.DB
}

func (p *Postgres) Open(ctx context.Context) error {
	db, err := sql.Open("pgx", p.dsn)
	if err != nil {
		return err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return err
	}
	p.db = db
	return nil
}

func (p *Postgres) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *Postgres) DB() *sql.DB { return p.db }

func (p *Postgres) Dialect() database.Dialect { return &Dialect{} }

func (p *Postgres) Lock(ctx context.Context) error {
	// Use pg_advisory_lock with a fixed lock ID derived from "drift"
	_, err := p.db.ExecContext(ctx, "SELECT pg_advisory_lock(2020817574)") // CRC32 of "drift"
	return err
}

func (p *Postgres) Unlock(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, "SELECT pg_advisory_unlock(2020817574)")
	return err
}

// Dialect implements database.Dialect for PostgreSQL.
type Dialect struct{}

func (d *Dialect) Name() string { return "postgres" }

func (d *Dialect) QuoteIdentifier(name string) string {
	return `"` + name + `"`
}

func (d *Dialect) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

func (d *Dialect) SupportsTransactionalDDL() bool { return true }

func (d *Dialect) CreateHistoryTableSQL(table string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    installed_rank INTEGER NOT NULL,
    version        VARCHAR(50),
    description    VARCHAR(200) NOT NULL,
    type           VARCHAR(20) NOT NULL,
    script         VARCHAR(1000) NOT NULL,
    checksum       INTEGER,
    installed_by   VARCHAR(100) NOT NULL,
    installed_host VARCHAR(255) NOT NULL DEFAULT '',
    installed_ip   VARCHAR(45) NOT NULL DEFAULT '',
    installed_on   TIMESTAMP NOT NULL DEFAULT now(),
    execution_time INTEGER NOT NULL,
    success        BOOLEAN NOT NULL,
    CONSTRAINT %s_pk PRIMARY KEY (installed_rank)
);
CREATE INDEX IF NOT EXISTS %s_s_idx ON %s (success);`,
		d.QuoteIdentifier(table),
		table,
		table, d.QuoteIdentifier(table))
}
