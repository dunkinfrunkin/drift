package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintNoDropTable(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "V001__bad.sql"), []byte("DROP TABLE users;"), 0644)

	l := NewLinter(nil)
	results, err := l.LintLocations([]string{dir})
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, r := range results {
		if r.Rule == "no-drop-table" {
			found = true
			if r.Severity != SeverityError {
				t.Errorf("expected ERROR severity, got %s", r.Severity)
			}
		}
	}
	if !found {
		t.Error("expected no-drop-table violation")
	}
}

func TestLintCleanFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "V001__create_users.sql"),
		[]byte("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);"), 0644)

	l := NewLinter(nil)
	results, err := l.LintLocations([]string{dir})
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 lint results for clean file, got %d: %v", len(results), results)
	}
}

func TestLintSelectStar(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "V001__view.sql"),
		[]byte("CREATE VIEW v AS SELECT * FROM users;"), 0644)

	l := NewLinter([]string{"no-select-star"})
	results, err := l.LintLocations([]string{dir})
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 || results[0].Rule != "no-select-star" {
		t.Errorf("expected 1 no-select-star result, got %v", results)
	}
}
