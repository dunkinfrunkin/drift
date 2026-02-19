package history

import (
	"context"
	"database/sql"
	"fmt"
	"os/user"
	"strings"
	"time"

	"github.com/frankchan/drift/internal/database"
	"github.com/frankchan/drift/internal/migration"
)

// History manages the drift_schema_history table.
type History struct {
	db    database.Database
	table string
}

// New creates a new History for the given database and table name.
func New(db database.Database, table string) *History {
	return &History{db: db, table: table}
}

// EnsureTable creates the history table if it doesn't exist.
func (h *History) EnsureTable(ctx context.Context) error {
	ddl := h.db.Dialect().CreateHistoryTableSQL(h.table)
	// Split on semicolons and execute each statement
	for _, stmt := range splitStatements(ddl) {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := h.db.DB().ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("creating history table: %w", err)
		}
	}
	return nil
}

// All returns all applied migrations ordered by installed_rank.
func (h *History) All(ctx context.Context) ([]migration.AppliedMigration, error) {
	qt := h.db.Dialect().QuoteIdentifier(h.table)
	rows, err := h.db.DB().QueryContext(ctx, fmt.Sprintf(
		`SELECT installed_rank, version, description, type, script, checksum,
		        installed_by, installed_on, execution_time, success
		 FROM %s ORDER BY installed_rank`, qt))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []migration.AppliedMigration
	for rows.Next() {
		var am migration.AppliedMigration
		var version sql.NullString
		var checksum sql.NullInt64
		var installedOn string
		if err := rows.Scan(
			&am.InstalledRank, &version, &am.Description, &am.Type, &am.Script,
			&checksum, &am.InstalledBy, &installedOn, &am.ExecutionTime, &am.Success,
		); err != nil {
			return nil, err
		}
		am.Version = version.String
		if checksum.Valid {
			am.Checksum = uint32(checksum.Int64)
		}
		am.InstalledOn, _ = time.Parse(time.RFC3339, installedOn)
		if am.InstalledOn.IsZero() {
			am.InstalledOn, _ = time.Parse("2006-01-02 15:04:05", installedOn)
		}
		result = append(result, am)
	}
	return result, rows.Err()
}

// Record inserts a new entry into the history table.
func (h *History) Record(ctx context.Context, m *migration.Migration, execTime time.Duration, success bool) error {
	d := h.db.Dialect()
	qt := d.QuoteIdentifier(h.table)

	nextRank, err := h.nextRank(ctx)
	if err != nil {
		return err
	}

	installedBy := currentUser()

	query := fmt.Sprintf(
		`INSERT INTO %s (installed_rank, version, description, type, script, checksum, installed_by, installed_on, execution_time, success)
		 VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)`,
		qt,
		d.Placeholder(1), d.Placeholder(2), d.Placeholder(3), d.Placeholder(4),
		d.Placeholder(5), d.Placeholder(6), d.Placeholder(7), d.Placeholder(8),
		d.Placeholder(9), d.Placeholder(10))

	_, err = h.db.DB().ExecContext(ctx, query,
		nextRank,
		nilIfEmpty(m.Version),
		m.Description,
		string(m.Type),
		m.Script,
		m.Checksum,
		installedBy,
		time.Now().UTC().Format(time.RFC3339),
		execTime.Milliseconds(),
		success,
	)
	return err
}

// Remove deletes history entries for a given version.
func (h *History) Remove(ctx context.Context, version string) error {
	d := h.db.Dialect()
	qt := d.QuoteIdentifier(h.table)
	_, err := h.db.DB().ExecContext(ctx,
		fmt.Sprintf("DELETE FROM %s WHERE version = %s", qt, d.Placeholder(1)),
		version)
	return err
}

// Clear deletes all rows from the history table.
func (h *History) Clear(ctx context.Context) error {
	qt := h.db.Dialect().QuoteIdentifier(h.table)
	_, err := h.db.DB().ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", qt))
	return err
}

// Drop removes the history table entirely.
func (h *History) Drop(ctx context.Context) error {
	qt := h.db.Dialect().QuoteIdentifier(h.table)
	_, err := h.db.DB().ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", qt))
	return err
}

// UpdateChecksum updates the checksum for a given version.
func (h *History) UpdateChecksum(ctx context.Context, version string, checksum uint32) error {
	d := h.db.Dialect()
	qt := d.QuoteIdentifier(h.table)
	_, err := h.db.DB().ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET checksum = %s WHERE version = %s", qt, d.Placeholder(1), d.Placeholder(2)),
		checksum, version)
	return err
}

// MarkSuccess updates the success flag for a given version.
func (h *History) MarkSuccess(ctx context.Context, version string, success bool) error {
	d := h.db.Dialect()
	qt := d.QuoteIdentifier(h.table)
	_, err := h.db.DB().ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET success = %s WHERE version = %s", qt, d.Placeholder(1), d.Placeholder(2)),
		success, version)
	return err
}

func (h *History) nextRank(ctx context.Context) (int, error) {
	qt := h.db.Dialect().QuoteIdentifier(h.table)
	var maxRank sql.NullInt64
	err := h.db.DB().QueryRowContext(ctx,
		fmt.Sprintf("SELECT MAX(installed_rank) FROM %s", qt)).Scan(&maxRank)
	if err != nil {
		return 1, nil
	}
	if !maxRank.Valid {
		return 1, nil
	}
	return int(maxRank.Int64) + 1, nil
}

func splitStatements(s string) []string {
	return strings.Split(s, ";")
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func currentUser() string {
	u, err := user.Current()
	if err != nil {
		return "drift"
	}
	return u.Username
}
