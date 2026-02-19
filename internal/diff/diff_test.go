package diff

import (
	"testing"
)

func TestCompareAddTable(t *testing.T) {
	from := &SchemaSnapshot{}
	to := &SchemaSnapshot{
		Tables: []Table{
			{Name: "users", Columns: []Column{
				{Name: "id", DataType: "INTEGER", Nullable: false},
				{Name: "name", DataType: "TEXT", Nullable: false},
			}},
		},
	}

	d := Compare(from, to)
	if len(d.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(d.Changes))
	}
	if d.Changes[0].Action != "ADD" || d.Changes[0].ObjectType != "TABLE" {
		t.Errorf("expected ADD TABLE, got %s %s", d.Changes[0].Action, d.Changes[0].ObjectType)
	}
}

func TestCompareDropTable(t *testing.T) {
	from := &SchemaSnapshot{
		Tables: []Table{{Name: "old_table", Columns: []Column{{Name: "id", DataType: "INT"}}}},
	}
	to := &SchemaSnapshot{}

	d := Compare(from, to)
	found := false
	for _, c := range d.Changes {
		if c.Action == "DROP" && c.ObjectType == "TABLE" && c.Name == "old_table" {
			found = true
		}
	}
	if !found {
		t.Error("expected DROP TABLE old_table")
	}
}

func TestCompareAddColumn(t *testing.T) {
	from := &SchemaSnapshot{
		Tables: []Table{{Name: "users", Columns: []Column{
			{Name: "id", DataType: "INTEGER"},
		}}},
	}
	to := &SchemaSnapshot{
		Tables: []Table{{Name: "users", Columns: []Column{
			{Name: "id", DataType: "INTEGER"},
			{Name: "email", DataType: "TEXT"},
		}}},
	}

	d := Compare(from, to)
	found := false
	for _, c := range d.Changes {
		if c.Action == "ADD" && c.ObjectType == "COLUMN" && c.Name == "users.email" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ADD COLUMN users.email, got %v", d.Changes)
	}
}

func TestCompareDropColumn(t *testing.T) {
	from := &SchemaSnapshot{
		Tables: []Table{{Name: "users", Columns: []Column{
			{Name: "id", DataType: "INTEGER"},
			{Name: "legacy", DataType: "TEXT"},
		}}},
	}
	to := &SchemaSnapshot{
		Tables: []Table{{Name: "users", Columns: []Column{
			{Name: "id", DataType: "INTEGER"},
		}}},
	}

	d := Compare(from, to)
	found := false
	for _, c := range d.Changes {
		if c.Action == "DROP" && c.ObjectType == "COLUMN" && c.Name == "users.legacy" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected DROP COLUMN users.legacy, got %v", d.Changes)
	}
}

func TestCompareModifyColumn(t *testing.T) {
	from := &SchemaSnapshot{
		Tables: []Table{{Name: "users", Columns: []Column{
			{Name: "name", DataType: "VARCHAR(50)", Nullable: false},
		}}},
	}
	to := &SchemaSnapshot{
		Tables: []Table{{Name: "users", Columns: []Column{
			{Name: "name", DataType: "VARCHAR(255)", Nullable: false},
		}}},
	}

	d := Compare(from, to)
	found := false
	for _, c := range d.Changes {
		if c.Action == "MODIFY" && c.Name == "users.name" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected MODIFY COLUMN users.name, got %v", d.Changes)
	}
}

func TestCompareNoChanges(t *testing.T) {
	snap := &SchemaSnapshot{
		Tables: []Table{{Name: "users", Columns: []Column{
			{Name: "id", DataType: "INTEGER"},
		}}},
	}

	d := Compare(snap, snap)
	if len(d.Changes) != 0 {
		t.Errorf("expected no changes, got %v", d.Changes)
	}
}

func TestCompareIndexChanges(t *testing.T) {
	from := &SchemaSnapshot{
		Indexes: []Index{{Name: "old_idx", Table: "users"}},
	}
	to := &SchemaSnapshot{
		Indexes: []Index{{Name: "new_idx", Table: "users"}},
	}

	d := Compare(from, to)
	var addFound, dropFound bool
	for _, c := range d.Changes {
		if c.Action == "ADD" && c.ObjectType == "INDEX" && c.Name == "new_idx" {
			addFound = true
		}
		if c.Action == "DROP" && c.ObjectType == "INDEX" && c.Name == "old_idx" {
			dropFound = true
		}
	}
	if !addFound {
		t.Error("expected ADD INDEX new_idx")
	}
	if !dropFound {
		t.Error("expected DROP INDEX old_idx")
	}
}

func TestFormatReport(t *testing.T) {
	d := &SchemaDiff{
		Changes: []Change{
			{Action: "ADD", ObjectType: "TABLE", Name: "users"},
			{Action: "DROP", ObjectType: "COLUMN", Name: "posts.legacy"},
		},
	}

	report := FormatReport(d)
	if report == "" {
		t.Error("expected non-empty report")
	}
}

func TestFormatReportEmpty(t *testing.T) {
	d := &SchemaDiff{}
	report := FormatReport(d)
	if report != "No changes detected." {
		t.Errorf("expected 'No changes detected.', got %q", report)
	}
}
