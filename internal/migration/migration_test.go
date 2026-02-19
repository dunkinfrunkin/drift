package migration

import (
	"testing"
)

func TestParseFilename(t *testing.T) {
	tests := []struct {
		filename string
		wantVer  string
		wantDesc string
		wantType Type
		wantErr  bool
	}{
		{"V001__create_users.sql", "001", "create users", TypeVersioned, false},
		{"V123__add_email_index.sql", "123", "add email index", TypeVersioned, false},
		{"R__refresh_views.sql", "", "refresh views", TypeRepeatable, false},
		{"U001__drop_users.sql", "001", "drop users", TypeUndo, false},
		{"beforeMigrate__audit.sql", "", "beforeMigrate: audit", TypeCallback, false},
		{"afterEachMigrate.sql", "", "afterEachMigrate", TypeCallback, false},
		{"random.sql", "", "", "", true},
		{"V__no_version.sql", "", "", "", true},
	}

	for _, tt := range tests {
		m, err := ParseFilename(tt.filename)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseFilename(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if m.Version != tt.wantVer {
			t.Errorf("ParseFilename(%q).Version = %q, want %q", tt.filename, m.Version, tt.wantVer)
		}
		if m.Description != tt.wantDesc {
			t.Errorf("ParseFilename(%q).Description = %q, want %q", tt.filename, m.Description, tt.wantDesc)
		}
		if m.Type != tt.wantType {
			t.Errorf("ParseFilename(%q).Type = %q, want %q", tt.filename, m.Type, tt.wantType)
		}
	}
}

func TestComputeChecksum(t *testing.T) {
	sql := "CREATE TABLE users (id INT);"
	c1 := ComputeChecksum(sql)
	c2 := ComputeChecksum(sql + "  \n")       // trailing whitespace
	c3 := ComputeChecksum("\r\n" + sql + "\r\n") // windows line endings + padding

	if c1 != c2 {
		t.Errorf("checksums should match after trimming: %d != %d", c1, c2)
	}
	if c1 != c3 {
		t.Errorf("checksums should match after CRLF normalization: %d != %d", c1, c3)
	}

	different := ComputeChecksum("DROP TABLE users;")
	if c1 == different {
		t.Error("different SQL should produce different checksums")
	}
}

func TestSubstitutePlaceholders(t *testing.T) {
	sql := "CREATE TABLE ${schema}.users (name ${varchar_type});"
	result := SubstitutePlaceholders(sql, map[string]string{
		"schema":       "public",
		"varchar_type": "VARCHAR(255)",
	})
	want := "CREATE TABLE public.users (name VARCHAR(255));"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}
