package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/frankchan/drift/internal/config"
	"github.com/frankchan/drift/internal/database"
	"github.com/frankchan/drift/internal/diff"
	"github.com/frankchan/drift/internal/history"
	"github.com/frankchan/drift/internal/migration"
)

// Engine is the core orchestrator for all drift operations.
type Engine struct {
	cfg      *config.Config
	db       database.Database
	history  *history.History
	resolver *Resolver
	planner  *Planner
	output   io.Writer
}

// New creates a new Engine.
func New(cfg *config.Config, db database.Database) *Engine {
	hist := history.New(db, cfg.Table)
	resolver := NewResolver(cfg.Locations, cfg.Placeholders)
	return &Engine{
		cfg:      cfg,
		db:       db,
		history:  hist,
		resolver: resolver,
		planner:  NewPlanner(),
		output:   os.Stdout,
	}
}

// SetOutput sets the writer for engine output messages.
func (e *Engine) SetOutput(w io.Writer) {
	e.output = w
}

// Migrate executes pending migrations.
func (e *Engine) Migrate(ctx context.Context, opts PlanOptions) ([]ExecuteResult, error) {
	if err := e.db.Lock(ctx); err != nil {
		return nil, fmt.Errorf("acquiring lock: %w", err)
	}
	defer e.db.Unlock(ctx)

	if err := e.history.EnsureTable(ctx); err != nil {
		return nil, err
	}

	// Run callbacks
	cbRunner := NewCallbackRunner(e.db, e.resolver, e.output)
	if err := cbRunner.Run(ctx, migration.BeforeMigrate); err != nil {
		return nil, err
	}

	discovered, err := e.resolver.Resolve()
	if err != nil {
		return nil, err
	}

	applied, err := e.history.All(ctx)
	if err != nil {
		return nil, err
	}

	plan, err := e.planner.PlanMigrate(discovered, applied, e.cfg.Baseline, opts)
	if err != nil {
		return nil, err
	}

	if len(plan.Migrations) == 0 {
		fmt.Fprintln(e.output, "Schema is up to date. No migrations to apply.")
		return nil, nil
	}

	// Capture pre-migration schema snapshot for auto-rollback
	var preSnapshot *diff.SchemaSnapshot
	if !opts.DryRun {
		preSnapshot, _ = diff.CaptureSchema(ctx, e.db)
	}

	executor := NewExecutor(e.db, e.history, e.output)
	var results []ExecuteResult

	for _, m := range plan.Migrations {
		fmt.Fprintf(e.output, "Migrating: %s\n", m.Script)
		if err := cbRunner.Run(ctx, migration.BeforeEach); err != nil {
			return results, err
		}

		result := executor.Execute(ctx, m, opts.DryRun)
		results = append(results, *result)

		if result.Success {
			fmt.Fprintf(e.output, "  Success  (%s)\n", result.ExecutionTime.Round(time.Millisecond))
		} else {
			fmt.Fprintf(e.output, "  FAILED   (%s)\n", result.ExecutionTime.Round(time.Millisecond))
			e.autoRollback(ctx, preSnapshot, results, executor)
			return results, result.Error
		}

		if err := cbRunner.Run(ctx, migration.AfterEach); err != nil {
			return results, err
		}
	}

	if err := cbRunner.Run(ctx, migration.AfterMigrate); err != nil {
		return results, err
	}

	fmt.Fprintf(e.output, "Successfully applied %d migration(s).\n", len(results))
	return results, nil
}

// autoRollback reverses successfully applied migrations from the current run after a failure.
func (e *Engine) autoRollback(ctx context.Context, preSnapshot *diff.SchemaSnapshot, results []ExecuteResult, executor *Executor) {
	// Collect successful versioned migrations (skip the failed one and non-versioned)
	var successful []ExecuteResult
	for _, r := range results {
		if r.Success && r.Migration.Type == migration.TypeVersioned {
			successful = append(successful, r)
		}
	}

	if len(successful) == 0 || preSnapshot == nil {
		return
	}

	fmt.Fprintf(e.output, "\nAuto-rolling back %d migration(s)...\n", len(successful))

	// Try undo files first
	undoByVersion := make(map[string]*migration.Migration)
	if undoFiles, err := e.resolver.ResolveByType(migration.TypeUndo); err == nil {
		for _, u := range undoFiles {
			undoByVersion[u.Version] = u
		}
	}

	// Rollback in reverse order
	for i := len(successful) - 1; i >= 0; i-- {
		r := successful[i]
		version := r.Migration.Version

		if undo, ok := undoByVersion[version]; ok {
			// Use the undo file
			fmt.Fprintf(e.output, "  Rolling back V%s (using %s)\n", version, undo.Script)
			result := executor.Execute(ctx, undo, false)
			if !result.Success {
				fmt.Fprintf(e.output, "  ROLLBACK FAILED for V%s: %v\n", version, result.Error)
				return
			}
		} else {
			// Generate reverse DDL from diff
			currentSchema, err := diff.CaptureSchema(ctx, e.db)
			if err != nil {
				fmt.Fprintf(e.output, "  ROLLBACK FAILED: could not capture schema: %v\n", err)
				return
			}
			changes := diff.Compare(preSnapshot, currentSchema)
			var rollbackSQL []string
			for _, c := range changes.Changes {
				if c.SQL != "" {
					rollbackSQL = append(rollbackSQL, c.SQL)
				}
			}
			if len(rollbackSQL) == 0 {
				fmt.Fprintf(e.output, "  Skipping V%s (no reverse DDL generated)\n", version)
				continue
			}

			fmt.Fprintf(e.output, "  Rolling back V%s (auto-generated)\n", version)
			rollbackM := &migration.Migration{
				Version:     version,
				Description: "auto-rollback",
				Type:        migration.TypeRollback,
				Script:      fmt.Sprintf("rollback_V%s", version),
				SQL:         strings.Join(rollbackSQL, ";\n"),
			}
			rollbackM.Checksum = migration.ComputeChecksum(rollbackM.SQL)
			result := executor.Execute(ctx, rollbackM, false)
			if !result.Success {
				fmt.Fprintf(e.output, "  ROLLBACK FAILED for V%s: %v\n", version, result.Error)
				return
			}
		}

		// Remove the forward migration's history entry
		if err := e.history.Remove(ctx, version); err != nil {
			fmt.Fprintf(e.output, "  Warning: could not remove history for V%s: %v\n", version, err)
		}
	}

	fmt.Fprintln(e.output, "Auto-rollback complete.")
}

// Undo reverses applied migrations.
func (e *Engine) Undo(ctx context.Context, count int, target string, dryRun bool) ([]ExecuteResult, error) {
	if err := e.db.Lock(ctx); err != nil {
		return nil, fmt.Errorf("acquiring lock: %w", err)
	}
	defer e.db.Unlock(ctx)

	if err := e.history.EnsureTable(ctx); err != nil {
		return nil, err
	}

	cbRunner := NewCallbackRunner(e.db, e.resolver, e.output)
	if err := cbRunner.Run(ctx, migration.BeforeUndo); err != nil {
		return nil, err
	}

	undoFiles, err := e.resolver.ResolveByType(migration.TypeUndo)
	if err != nil {
		return nil, err
	}

	applied, err := e.history.All(ctx)
	if err != nil {
		return nil, err
	}

	if count == 0 && target == "" {
		count = 1
	}

	plan, err := e.planner.PlanUndo(undoFiles, applied, count, target)
	if err != nil {
		return nil, err
	}

	if len(plan.Migrations) == 0 {
		fmt.Fprintln(e.output, "Nothing to undo.")
		return nil, nil
	}

	executor := NewExecutor(e.db, e.history, e.output)
	var results []ExecuteResult

	for _, m := range plan.Migrations {
		fmt.Fprintf(e.output, "Undoing: %s\n", m.Script)
		result := executor.Execute(ctx, m, dryRun)
		results = append(results, *result)

		if result.Success {
			if !dryRun {
				// Remove the forward migration's history entry
				if err := e.history.Remove(ctx, m.Version); err != nil {
					return results, fmt.Errorf("removing history for V%s: %w", m.Version, err)
				}
			}
			fmt.Fprintf(e.output, "  Success  (%s)\n", result.ExecutionTime.Round(time.Millisecond))
		} else {
			fmt.Fprintf(e.output, "  FAILED   (%s)\n", result.ExecutionTime.Round(time.Millisecond))
			return results, result.Error
		}
	}

	if err := cbRunner.Run(ctx, migration.AfterUndo); err != nil {
		return results, err
	}

	fmt.Fprintf(e.output, "Successfully undone %d migration(s).\n", len(results))
	return results, nil
}

// Info returns the current migration status.
func (e *Engine) Info(ctx context.Context) ([]MigrationInfo, error) {
	if err := e.history.EnsureTable(ctx); err != nil {
		return nil, err
	}

	discovered, err := e.resolver.Resolve()
	if err != nil {
		return nil, err
	}

	applied, err := e.history.All(ctx)
	if err != nil {
		return nil, err
	}

	return buildInfo(discovered, applied), nil
}

// MigrationInfo combines discovered and applied migration data.
type MigrationInfo struct {
	Version     string
	Description string
	Type        migration.Type
	Script      string
	State       migration.State
	InstalledOn time.Time
	ExecTime    time.Duration
}

func buildInfo(discovered []*migration.Migration, applied []migration.AppliedMigration) []MigrationInfo {
	appliedMap := make(map[string]migration.AppliedMigration)
	for _, a := range applied {
		appliedMap[a.Version] = a
	}

	var infos []MigrationInfo

	// Add all discovered versioned migrations
	for _, m := range discovered {
		if m.Type != migration.TypeVersioned {
			continue
		}
		info := MigrationInfo{
			Version:     m.Version,
			Description: m.Description,
			Type:        m.Type,
			Script:      m.Script,
			State:       migration.StatePending,
		}
		if a, ok := appliedMap[m.Version]; ok {
			if a.Success {
				info.State = migration.StateApplied
			} else {
				info.State = migration.StateFailed
			}
			info.InstalledOn = a.InstalledOn
			info.ExecTime = time.Duration(a.ExecutionTime) * time.Millisecond
			delete(appliedMap, m.Version)
		}
		infos = append(infos, info)
	}

	// Add applied migrations that are no longer on filesystem
	for _, a := range applied {
		if _, found := appliedMap[a.Version]; !found {
			continue
		}
		if a.Type != migration.TypeVersioned {
			continue
		}
		infos = append(infos, MigrationInfo{
			Version:     a.Version,
			Description: a.Description,
			Type:        a.Type,
			Script:      a.Script,
			State:       migration.State("Missing"),
			InstalledOn: a.InstalledOn,
			ExecTime:    time.Duration(a.ExecutionTime) * time.Millisecond,
		})
	}

	return infos
}

// Validate checks that applied migrations match the filesystem.
func (e *Engine) Validate(ctx context.Context) ([]ValidationError, error) {
	if err := e.history.EnsureTable(ctx); err != nil {
		return nil, err
	}

	discovered, err := e.resolver.Resolve()
	if err != nil {
		return nil, err
	}

	applied, err := e.history.All(ctx)
	if err != nil {
		return nil, err
	}

	v := NewValidator()
	return v.Validate(discovered, applied), nil
}

// Clean drops all objects in the configured schemas.
func (e *Engine) Clean(ctx context.Context) error {
	if err := e.db.Lock(ctx); err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	defer e.db.Unlock(ctx)

	cbRunner := NewCallbackRunner(e.db, e.resolver, e.output)
	if err := cbRunner.Run(ctx, migration.BeforeClean); err != nil {
		return err
	}

	// Drop the history table
	hist := history.New(e.db, e.cfg.Table)
	if err := hist.Drop(ctx); err != nil {
		return fmt.Errorf("dropping history table: %w", err)
	}

	fmt.Fprintln(e.output, "Schema history table dropped.")

	if err := cbRunner.Run(ctx, migration.AfterClean); err != nil {
		return err
	}

	return nil
}

// Baseline marks a version as the baseline.
func (e *Engine) Baseline(ctx context.Context, version string) error {
	if err := e.db.Lock(ctx); err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	defer e.db.Unlock(ctx)

	if err := e.history.EnsureTable(ctx); err != nil {
		return err
	}

	cbRunner := NewCallbackRunner(e.db, e.resolver, e.output)
	if err := cbRunner.Run(ctx, migration.BeforeBaseline); err != nil {
		return err
	}

	m := &migration.Migration{
		Version:     version,
		Description: "<< Baseline >>",
		Type:        migration.TypeVersioned,
		Script:      "<< Baseline >>",
		Checksum:    0,
	}

	if err := e.history.Record(ctx, m, 0, true); err != nil {
		return fmt.Errorf("recording baseline: %w", err)
	}

	fmt.Fprintf(e.output, "Baselined at version V%s.\n", version)

	if err := cbRunner.Run(ctx, migration.AfterBaseline); err != nil {
		return err
	}

	return nil
}

// Repair fixes the schema history table by realigning checksums.
func (e *Engine) Repair(ctx context.Context) error {
	if err := e.db.Lock(ctx); err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	defer e.db.Unlock(ctx)

	if err := e.history.EnsureTable(ctx); err != nil {
		return err
	}

	cbRunner := NewCallbackRunner(e.db, e.resolver, e.output)
	if err := cbRunner.Run(ctx, migration.BeforeRepair); err != nil {
		return err
	}

	discovered, err := e.resolver.Resolve()
	if err != nil {
		return err
	}

	applied, err := e.history.All(ctx)
	if err != nil {
		return err
	}

	discoveredMap := make(map[string]*migration.Migration)
	for _, m := range discovered {
		if m.Type == migration.TypeVersioned {
			discoveredMap[m.Version] = m
		}
	}

	repaired := 0
	for _, a := range applied {
		if !a.Success {
			// Remove failed entries
			if err := e.history.Remove(ctx, a.Version); err != nil {
				return fmt.Errorf("removing failed entry V%s: %w", a.Version, err)
			}
			fmt.Fprintf(e.output, "Removed failed migration: V%s\n", a.Version)
			repaired++
			continue
		}

		if m, ok := discoveredMap[a.Version]; ok && m.Checksum != a.Checksum {
			if err := e.history.UpdateChecksum(ctx, a.Version, m.Checksum); err != nil {
				return fmt.Errorf("updating checksum for V%s: %w", a.Version, err)
			}
			fmt.Fprintf(e.output, "Repaired checksum: V%s\n", a.Version)
			repaired++
		}
	}

	if repaired == 0 {
		fmt.Fprintln(e.output, "Nothing to repair.")
	} else {
		fmt.Fprintf(e.output, "Repaired %d migration(s).\n", repaired)
	}

	if err := cbRunner.Run(ctx, migration.AfterRepair); err != nil {
		return err
	}

	return nil
}

// History returns the underlying history manager.
func (e *Engine) History() *history.History {
	return e.history
}

// Resolver returns the underlying resolver.
func (e *Engine) Resolver() *Resolver {
	return e.resolver
}
