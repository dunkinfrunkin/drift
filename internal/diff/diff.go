package diff

import "fmt"

// SchemaDiff represents the differences between two schema snapshots.
type SchemaDiff struct {
	Changes []Change `json:"changes"`
}

// Change represents a single schema change.
type Change struct {
	Action     string `json:"action"` // ADD, DROP, MODIFY
	ObjectType string `json:"objectType"` // TABLE, COLUMN, INDEX
	Name       string `json:"name"`
	Details    string `json:"details,omitempty"`
	SQL        string `json:"sql,omitempty"`
}

// Compare computes the diff between two schema snapshots (from → to).
func Compare(from, to *SchemaSnapshot) *SchemaDiff {
	diff := &SchemaDiff{}

	fromTables := indexTables(from.Tables)
	toTables := indexTables(to.Tables)

	// Find added and modified tables
	for name, toTable := range toTables {
		fromTable, exists := fromTables[name]
		if !exists {
			diff.Changes = append(diff.Changes, Change{
				Action:     "ADD",
				ObjectType: "TABLE",
				Name:       name,
				SQL:        generateCreateTableSQL(toTable),
			})
			continue
		}

		// Compare columns
		compareCols(diff, name, fromTable.Columns, toTable.Columns)
	}

	// Find dropped tables
	for name := range fromTables {
		if _, exists := toTables[name]; !exists {
			diff.Changes = append(diff.Changes, Change{
				Action:     "DROP",
				ObjectType: "TABLE",
				Name:       name,
				SQL:        fmt.Sprintf("DROP TABLE %s", name),
			})
		}
	}

	// Compare indexes
	fromIdxs := indexIndexes(from.Indexes)
	toIdxs := indexIndexes(to.Indexes)

	for name := range toIdxs {
		if _, exists := fromIdxs[name]; !exists {
			diff.Changes = append(diff.Changes, Change{
				Action:     "ADD",
				ObjectType: "INDEX",
				Name:       name,
			})
		}
	}

	for name := range fromIdxs {
		if _, exists := toIdxs[name]; !exists {
			diff.Changes = append(diff.Changes, Change{
				Action:     "DROP",
				ObjectType: "INDEX",
				Name:       name,
				SQL:        fmt.Sprintf("DROP INDEX %s", name),
			})
		}
	}

	return diff
}

func compareCols(diff *SchemaDiff, tableName string, from, to []Column) {
	fromCols := make(map[string]Column)
	for _, c := range from {
		fromCols[c.Name] = c
	}

	toCols := make(map[string]Column)
	for _, c := range to {
		toCols[c.Name] = c
	}

	for name, toCol := range toCols {
		fromCol, exists := fromCols[name]
		if !exists {
			diff.Changes = append(diff.Changes, Change{
				Action:     "ADD",
				ObjectType: "COLUMN",
				Name:       fmt.Sprintf("%s.%s", tableName, name),
				Details:    toCol.DataType,
			})
			continue
		}
		if fromCol.DataType != toCol.DataType || fromCol.Nullable != toCol.Nullable {
			diff.Changes = append(diff.Changes, Change{
				Action:     "MODIFY",
				ObjectType: "COLUMN",
				Name:       fmt.Sprintf("%s.%s", tableName, name),
				Details:    fmt.Sprintf("%s → %s", fromCol.DataType, toCol.DataType),
			})
		}
	}

	for name := range fromCols {
		if _, exists := toCols[name]; !exists {
			diff.Changes = append(diff.Changes, Change{
				Action:     "DROP",
				ObjectType: "COLUMN",
				Name:       fmt.Sprintf("%s.%s", tableName, name),
			})
		}
	}
}

func generateCreateTableSQL(t Table) string {
	sql := fmt.Sprintf("CREATE TABLE %s (\n", t.Name)
	for i, c := range t.Columns {
		nullable := " NOT NULL"
		if c.Nullable {
			nullable = ""
		}
		sql += fmt.Sprintf("    %s %s%s", c.Name, c.DataType, nullable)
		if i < len(t.Columns)-1 {
			sql += ","
		}
		sql += "\n"
	}
	sql += ")"
	return sql
}

func indexTables(tables []Table) map[string]Table {
	m := make(map[string]Table, len(tables))
	for _, t := range tables {
		m[t.Name] = t
	}
	return m
}

func indexIndexes(indexes []Index) map[string]Index {
	m := make(map[string]Index, len(indexes))
	for _, idx := range indexes {
		m[idx.Name] = idx
	}
	return m
}
