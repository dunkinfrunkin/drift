package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/frankchan/drift/internal/database"
)

// CaptureSchema introspects the database and returns a schema snapshot.
func CaptureSchema(ctx context.Context, db database.Database) (*SchemaSnapshot, error) {
	dialect := db.Dialect().Name()
	switch dialect {
	case "postgres":
		return capturePostgres(ctx, db)
	case "mysql":
		return captureMySQL(ctx, db)
	case "sqlite":
		return captureSQLite(ctx, db)
	default:
		return nil, fmt.Errorf("schema introspection not supported for %s", dialect)
	}
}

func capturePostgres(ctx context.Context, db database.Database) (*SchemaSnapshot, error) {
	snap := &SchemaSnapshot{}

	// Get tables
	rows, err := db.DB().QueryContext(ctx, `
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		  AND table_type = 'BASE TABLE'
		ORDER BY table_schema, table_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Table
		if err := rows.Scan(&t.Schema, &t.Name); err != nil {
			return nil, err
		}

		// Get columns
		cols, err := db.DB().QueryContext(ctx, `
			SELECT column_name, data_type, is_nullable, COALESCE(column_default, '')
			FROM information_schema.columns
			WHERE table_schema = $1 AND table_name = $2
			ORDER BY ordinal_position`, t.Schema, t.Name)
		if err != nil {
			return nil, err
		}

		for cols.Next() {
			var c Column
			var nullable string
			if err := cols.Scan(&c.Name, &c.DataType, &nullable, &c.DefaultValue); err != nil {
				cols.Close()
				return nil, err
			}
			c.Nullable = nullable == "YES"
			t.Columns = append(t.Columns, c)
		}
		cols.Close()

		snap.Tables = append(snap.Tables, t)
	}

	// Get indexes
	idxRows, err := db.DB().QueryContext(ctx, `
		SELECT indexname, tablename, indexdef
		FROM pg_indexes
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY tablename, indexname`)
	if err != nil {
		return nil, err
	}
	defer idxRows.Close()

	for idxRows.Next() {
		var idx Index
		var indexdef string
		if err := idxRows.Scan(&idx.Name, &idx.Table, &indexdef); err != nil {
			return nil, err
		}
		snap.Indexes = append(snap.Indexes, idx)
	}

	return snap, nil
}

func captureMySQL(ctx context.Context, db database.Database) (*SchemaSnapshot, error) {
	snap := &SchemaSnapshot{}

	rows, err := db.DB().QueryContext(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Table
		if err := rows.Scan(&t.Name); err != nil {
			return nil, err
		}

		cols, err := db.DB().QueryContext(ctx, `
			SELECT column_name, data_type, is_nullable, COALESCE(column_default, '')
			FROM information_schema.columns
			WHERE table_schema = DATABASE() AND table_name = ?
			ORDER BY ordinal_position`, t.Name)
		if err != nil {
			return nil, err
		}

		for cols.Next() {
			var c Column
			var nullable string
			if err := cols.Scan(&c.Name, &c.DataType, &nullable, &c.DefaultValue); err != nil {
				cols.Close()
				return nil, err
			}
			c.Nullable = nullable == "YES"
			t.Columns = append(t.Columns, c)
		}
		cols.Close()

		snap.Tables = append(snap.Tables, t)
	}

	return snap, nil
}

func captureSQLite(ctx context.Context, db database.Database) (*SchemaSnapshot, error) {
	snap := &SchemaSnapshot{}

	rows, err := db.DB().QueryContext(ctx, `
		SELECT name FROM sqlite_master
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Table
		if err := rows.Scan(&t.Name); err != nil {
			return nil, err
		}

		cols, err := db.DB().QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%q)", t.Name))
		if err != nil {
			return nil, err
		}

		for cols.Next() {
			var cid int
			var c Column
			var notNull int
			var dflt *string
			var pk int
			if err := cols.Scan(&cid, &c.Name, &c.DataType, &notNull, &dflt, &pk); err != nil {
				cols.Close()
				return nil, err
			}
			c.Nullable = notNull == 0
			if dflt != nil {
				c.DefaultValue = *dflt
			}
			if pk > 0 {
				t.PrimaryKey = append(t.PrimaryKey, c.Name)
			}
			t.Columns = append(t.Columns, c)
		}
		cols.Close()

		snap.Tables = append(snap.Tables, t)
	}

	// Get indexes
	idxRows, err := db.DB().QueryContext(ctx, `
		SELECT name, tbl_name FROM sqlite_master
		WHERE type = 'index' AND name NOT LIKE 'sqlite_%'
		ORDER BY tbl_name, name`)
	if err != nil {
		return nil, err
	}
	defer idxRows.Close()

	for idxRows.Next() {
		var idx Index
		if err := idxRows.Scan(&idx.Name, &idx.Table); err != nil {
			return nil, err
		}
		snap.Indexes = append(snap.Indexes, idx)
	}

	return snap, nil
}

// LoadSnapshot loads a SchemaSnapshot from a JSON file.
func LoadSnapshot(path string) (*SchemaSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap SchemaSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}
