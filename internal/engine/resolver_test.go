package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/frankchan/drift/internal/migration"
)

func TestResolve(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "V002__second.sql"), []byte("SELECT 2;"), 0644)
	os.WriteFile(filepath.Join(dir, "V001__first.sql"), []byte("SELECT 1;"), 0644)
	os.WriteFile(filepath.Join(dir, "R__views.sql"), []byte("SELECT 3;"), 0644)
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not a migration"), 0644)

	r := NewResolver([]string{dir}, nil)
	all, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}

	// Should find 3 SQL files, sorted (R_ sorts before V_)
	if len(all) != 3 {
		t.Fatalf("expected 3 migrations, got %d", len(all))
	}
	if all[0].Type != migration.TypeRepeatable {
		t.Errorf("first should be repeatable (R_ < V_), got %s", all[0].Type)
	}
	if all[1].Version != "001" {
		t.Errorf("second should be V001, got V%s", all[1].Version)
	}
	if all[2].Version != "002" {
		t.Errorf("third should be V002, got V%s", all[2].Version)
	}
}

func TestResolveMultipleLocations(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	os.WriteFile(filepath.Join(dir1, "V001__a.sql"), []byte("SELECT 1;"), 0644)
	os.WriteFile(filepath.Join(dir2, "V002__b.sql"), []byte("SELECT 2;"), 0644)

	r := NewResolver([]string{dir1, dir2}, nil)
	all, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2, got %d", len(all))
	}
}

func TestResolveWithPlaceholders(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "V001__init.sql"), []byte("CREATE TABLE ${schema}.users (id INT);"), 0644)

	r := NewResolver([]string{dir}, map[string]string{"schema": "myapp"})
	all, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatal("expected 1")
	}
	if all[0].SQL != "CREATE TABLE myapp.users (id INT);" {
		t.Errorf("placeholder not substituted: %q", all[0].SQL)
	}
}

func TestResolveMissingDir(t *testing.T) {
	r := NewResolver([]string{"/nonexistent/path"}, nil)
	all, err := r.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 from missing dir, got %d", len(all))
	}
}

func TestResolveByType(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "V001__a.sql"), []byte("SELECT 1;"), 0644)
	os.WriteFile(filepath.Join(dir, "U001__a.sql"), []byte("DROP;"), 0644)
	os.WriteFile(filepath.Join(dir, "R__views.sql"), []byte("SELECT 3;"), 0644)

	r := NewResolver([]string{dir}, nil)
	undos, err := r.ResolveByType(migration.TypeUndo)
	if err != nil {
		t.Fatal(err)
	}
	if len(undos) != 1 {
		t.Fatalf("expected 1 undo, got %d", len(undos))
	}
}

func TestResolveCallbacks(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "beforeMigrate__audit.sql"), []byte("INSERT INTO log VALUES ('start');"), 0644)
	os.WriteFile(filepath.Join(dir, "afterMigrate__notify.sql"), []byte("NOTIFY done;"), 0644)
	os.WriteFile(filepath.Join(dir, "V001__a.sql"), []byte("SELECT 1;"), 0644)

	r := NewResolver([]string{dir}, nil)
	cbs, err := r.ResolveCallbacks("beforeMigrate")
	if err != nil {
		t.Fatal(err)
	}
	if len(cbs) != 1 {
		t.Fatalf("expected 1 beforeMigrate callback, got %d", len(cbs))
	}
}
