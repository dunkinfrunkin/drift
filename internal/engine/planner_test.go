package engine

import (
	"testing"

	"github.com/frankchan/drift/internal/migration"
)

func TestPlanMigrateNoPending(t *testing.T) {
	p := NewPlanner()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned, Script: "V001__a.sql", Checksum: 1},
	}
	applied := []migration.AppliedMigration{
		{Version: "001", Type: migration.TypeVersioned, Script: "V001__a.sql", Checksum: 1, Success: true},
	}

	plan, err := p.PlanMigrate(discovered, applied, "", PlanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Migrations) != 0 {
		t.Errorf("expected 0 pending, got %d", len(plan.Migrations))
	}
}

func TestPlanMigrateWithPending(t *testing.T) {
	p := NewPlanner()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned, Script: "V001__a.sql", Checksum: 1},
		{Version: "002", Type: migration.TypeVersioned, Script: "V002__b.sql", Checksum: 2},
		{Version: "003", Type: migration.TypeVersioned, Script: "V003__c.sql", Checksum: 3},
	}
	applied := []migration.AppliedMigration{
		{Version: "001", Type: migration.TypeVersioned, Checksum: 1, Success: true},
	}

	plan, err := p.PlanMigrate(discovered, applied, "", PlanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Migrations) != 2 {
		t.Fatalf("expected 2 pending, got %d", len(plan.Migrations))
	}
	if plan.Migrations[0].Version != "002" || plan.Migrations[1].Version != "003" {
		t.Errorf("unexpected order: %s, %s", plan.Migrations[0].Version, plan.Migrations[1].Version)
	}
}

func TestPlanMigrateWithTarget(t *testing.T) {
	p := NewPlanner()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned},
		{Version: "002", Type: migration.TypeVersioned},
		{Version: "003", Type: migration.TypeVersioned},
	}

	plan, err := p.PlanMigrate(discovered, nil, "", PlanOptions{Target: "002"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Migrations) != 2 {
		t.Fatalf("expected 2 with target, got %d", len(plan.Migrations))
	}
}

func TestPlanMigrateWithSkip(t *testing.T) {
	p := NewPlanner()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned},
		{Version: "002", Type: migration.TypeVersioned},
		{Version: "003", Type: migration.TypeVersioned},
	}

	plan, err := p.PlanMigrate(discovered, nil, "", PlanOptions{Skip: []string{"002"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Migrations) != 2 {
		t.Fatalf("expected 2 after skip, got %d", len(plan.Migrations))
	}
	for _, m := range plan.Migrations {
		if m.Version == "002" {
			t.Error("version 002 should have been skipped")
		}
	}
}

func TestPlanMigrateCherryPick(t *testing.T) {
	p := NewPlanner()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned},
		{Version: "002", Type: migration.TypeVersioned},
		{Version: "003", Type: migration.TypeVersioned},
	}

	plan, err := p.PlanMigrate(discovered, nil, "", PlanOptions{CherryPick: []string{"001", "003"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Migrations) != 2 {
		t.Fatalf("expected 2 cherry-picked, got %d", len(plan.Migrations))
	}
}

func TestPlanMigrateBaselineSkipsOld(t *testing.T) {
	p := NewPlanner()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned},
		{Version: "002", Type: migration.TypeVersioned},
		{Version: "003", Type: migration.TypeVersioned},
	}

	plan, err := p.PlanMigrate(discovered, nil, "002", PlanOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Migrations) != 1 {
		t.Fatalf("expected 1 after baseline, got %d", len(plan.Migrations))
	}
	if plan.Migrations[0].Version != "003" {
		t.Errorf("expected V003, got V%s", plan.Migrations[0].Version)
	}
}

func TestPlanMigrateOutOfOrderError(t *testing.T) {
	p := NewPlanner()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned},
		{Version: "003", Type: migration.TypeVersioned},
	}
	applied := []migration.AppliedMigration{
		{Version: "003", Type: migration.TypeVersioned, Success: true},
	}

	_, err := p.PlanMigrate(discovered, applied, "", PlanOptions{})
	if err == nil {
		t.Error("expected out-of-order error")
	}
}

func TestPlanMigrateOutOfOrderAllowed(t *testing.T) {
	p := NewPlanner()
	discovered := []*migration.Migration{
		{Version: "001", Type: migration.TypeVersioned},
		{Version: "003", Type: migration.TypeVersioned},
	}
	applied := []migration.AppliedMigration{
		{Version: "003", Type: migration.TypeVersioned, Success: true},
	}

	plan, err := p.PlanMigrate(discovered, applied, "", PlanOptions{OutOfOrder: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Migrations) != 1 {
		t.Fatalf("expected 1 out-of-order, got %d", len(plan.Migrations))
	}
}

func TestPlanMigrateRepeatable(t *testing.T) {
	p := NewPlanner()
	discovered := []*migration.Migration{
		{Type: migration.TypeRepeatable, Script: "R__views.sql", Checksum: 100, Description: "views"},
	}

	// First run: should be pending
	plan, _ := p.PlanMigrate(discovered, nil, "", PlanOptions{})
	if len(plan.Migrations) != 1 {
		t.Fatalf("expected 1 repeatable, got %d", len(plan.Migrations))
	}

	// Second run with same checksum: should skip
	applied := []migration.AppliedMigration{
		{Script: "R__views.sql", Checksum: 100, Success: true, Type: migration.TypeRepeatable},
	}
	plan, _ = p.PlanMigrate(discovered, applied, "", PlanOptions{})
	if len(plan.Migrations) != 0 {
		t.Fatalf("expected 0 when checksum matches, got %d", len(plan.Migrations))
	}

	// Third run with changed checksum: should re-run
	discovered[0].Checksum = 200
	plan, _ = p.PlanMigrate(discovered, applied, "", PlanOptions{})
	if len(plan.Migrations) != 1 {
		t.Fatalf("expected 1 when checksum changed, got %d", len(plan.Migrations))
	}
}

func TestPlanUndo(t *testing.T) {
	p := NewPlanner()
	undoFiles := []*migration.Migration{
		{Version: "001", Type: migration.TypeUndo, Script: "U001__a.sql"},
		{Version: "002", Type: migration.TypeUndo, Script: "U002__b.sql"},
	}
	applied := []migration.AppliedMigration{
		{Version: "001", Type: migration.TypeVersioned, Success: true},
		{Version: "002", Type: migration.TypeVersioned, Success: true},
	}

	plan, err := p.PlanUndo(undoFiles, applied, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Migrations) != 1 {
		t.Fatalf("expected 1 undo, got %d", len(plan.Migrations))
	}
	if plan.Migrations[0].Version != "002" {
		t.Errorf("expected V002 undo, got V%s", plan.Migrations[0].Version)
	}
}

func TestPlanUndoToTarget(t *testing.T) {
	p := NewPlanner()
	undoFiles := []*migration.Migration{
		{Version: "001", Type: migration.TypeUndo},
		{Version: "002", Type: migration.TypeUndo},
		{Version: "003", Type: migration.TypeUndo},
	}
	applied := []migration.AppliedMigration{
		{Version: "001", Type: migration.TypeVersioned, Success: true},
		{Version: "002", Type: migration.TypeVersioned, Success: true},
		{Version: "003", Type: migration.TypeVersioned, Success: true},
	}

	plan, err := p.PlanUndo(undoFiles, applied, 0, "001")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Migrations) != 2 {
		t.Fatalf("expected 2 undos to target, got %d", len(plan.Migrations))
	}
}

func TestPlanUndoMissingFile(t *testing.T) {
	p := NewPlanner()
	applied := []migration.AppliedMigration{
		{Version: "001", Type: migration.TypeVersioned, Success: true},
	}

	_, err := p.PlanUndo(nil, applied, 1, "")
	if err == nil {
		t.Error("expected error for missing undo file")
	}
}
