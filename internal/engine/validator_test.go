package engine

import (
	"testing"

	"github.com/frankchan/drift/internal/migration"
)

func TestValidateClean(t *testing.T) {
	v := NewValidator()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned, Checksum: 100},
	}
	applied := []migration.AppliedMigration{
		{Version: "001", Type: migration.TypeVersioned, Checksum: 100, Success: true},
	}

	errs := v.Validate(discovered, applied)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateChecksumMismatch(t *testing.T) {
	v := NewValidator()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned, Checksum: 200},
	}
	applied := []migration.AppliedMigration{
		{Version: "001", Type: migration.TypeVersioned, Checksum: 100, Success: true},
	}

	errs := v.Validate(discovered, applied)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Version != "001" {
		t.Errorf("expected version 001, got %s", errs[0].Version)
	}
}

func TestValidateMissingFile(t *testing.T) {
	v := NewValidator()
	discovered := []*migration.Migration{} // file deleted
	applied := []migration.AppliedMigration{
		{Version: "001", Type: migration.TypeVersioned, Checksum: 100, Success: true},
	}

	errs := v.Validate(discovered, applied)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestValidateFailedEntry(t *testing.T) {
	v := NewValidator()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned, Checksum: 100},
	}
	applied := []migration.AppliedMigration{
		{Version: "001", Type: migration.TypeVersioned, Checksum: 100, Success: false},
	}

	errs := v.Validate(discovered, applied)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for failed entry, got %d", len(errs))
	}
}
