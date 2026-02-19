package engine

import (
	"fmt"

	"github.com/frankchan/drift/internal/migration"
)

// Plan represents the ordered list of migrations to execute.
type Plan struct {
	Migrations []*migration.Migration
	Direction  Direction
}

// Direction indicates whether we're migrating forward or undoing.
type Direction string

const (
	DirectionUp   Direction = "up"
	DirectionDown Direction = "down"
)

// PlanOptions configures how the planner builds the execution plan.
type PlanOptions struct {
	Target     string   // stop at this version
	CherryPick []string // only apply these versions
	Skip       []string // skip these versions
	OutOfOrder bool     // allow applying out-of-order migrations
	DryRun     bool     // don't actually execute
}

// Planner determines which migrations need to be applied.
type Planner struct{}

// NewPlanner creates a new Planner.
func NewPlanner() *Planner {
	return &Planner{}
}

// PlanMigrate builds a plan for forward migration.
func (p *Planner) PlanMigrate(
	discovered []*migration.Migration,
	applied []migration.AppliedMigration,
	baseline string,
	opts PlanOptions,
) (*Plan, error) {
	appliedVersions := make(map[string]migration.AppliedMigration)
	for _, a := range applied {
		if a.Success {
			appliedVersions[a.Version] = a
		}
	}

	skipSet := toSet(opts.Skip)
	cherrySet := toSet(opts.CherryPick)
	hasCherryPick := len(cherrySet) > 0

	var pending []*migration.Migration

	for _, m := range discovered {
		switch m.Type {
		case migration.TypeVersioned:
			// Skip versions at or below baseline
			if baseline != "" && m.Version <= baseline {
				continue
			}
			// Skip already applied
			if _, ok := appliedVersions[m.Version]; ok {
				continue
			}
			// Skip if in skip list
			if skipSet[m.Version] {
				continue
			}
			// Cherry-pick filter
			if hasCherryPick && !cherrySet[m.Version] {
				continue
			}
			// Check out-of-order
			if !opts.OutOfOrder && len(applied) > 0 {
				lastApplied := lastSuccessfulVersion(applied)
				if lastApplied != "" && m.Version < lastApplied {
					return nil, fmt.Errorf("migration V%s is older than last applied V%s; use --out-of-order to allow", m.Version, lastApplied)
				}
			}
			// Target filter
			if opts.Target != "" && m.Version > opts.Target {
				continue
			}
			pending = append(pending, m)

		case migration.TypeRepeatable:
			// Repeatable migrations run if checksum changed or not yet applied
			found := false
			for _, a := range applied {
				if a.Script == m.Script {
					if a.Checksum == m.Checksum && a.Success {
						found = true
					}
					break
				}
			}
			if !found {
				pending = append(pending, m)
			}
		}
	}

	return &Plan{Migrations: pending, Direction: DirectionUp}, nil
}

// PlanUndo builds a plan for undoing migrations.
func (p *Planner) PlanUndo(
	undoFiles []*migration.Migration,
	applied []migration.AppliedMigration,
	count int,
	target string,
) (*Plan, error) {
	// Build a map of undo files by version
	undoByVersion := make(map[string]*migration.Migration)
	for _, u := range undoFiles {
		undoByVersion[u.Version] = u
	}

	// Get successfully applied versioned migrations in reverse order
	var toUndo []*migration.Migration
	for i := len(applied) - 1; i >= 0; i-- {
		a := applied[i]
		if !a.Success || a.Type != migration.TypeVersioned {
			continue
		}
		if target != "" && a.Version <= target {
			break
		}
		u, ok := undoByVersion[a.Version]
		if !ok {
			return nil, fmt.Errorf("no undo migration found for V%s", a.Version)
		}
		toUndo = append(toUndo, u)
		if count > 0 && len(toUndo) >= count {
			break
		}
	}

	return &Plan{Migrations: toUndo, Direction: DirectionDown}, nil
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}

func lastSuccessfulVersion(applied []migration.AppliedMigration) string {
	for i := len(applied) - 1; i >= 0; i-- {
		if applied[i].Success && applied[i].Type == migration.TypeVersioned {
			return applied[i].Version
		}
	}
	return ""
}
