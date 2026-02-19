package migration

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Type represents the kind of migration.
type Type string

const (
	TypeVersioned   Type = "VERSIONED"
	TypeRepeatable  Type = "REPEATABLE"
	TypeUndo        Type = "UNDO"
	TypeCallback    Type = "CALLBACK"
)

// State represents the execution state of a migration.
type State string

const (
	StatePending       State = "Pending"
	StateApplied       State = "Applied"
	StateFailed        State = "Failed"
	StateOutdated      State = "Outdated"
	StateUndone        State = "Undone"
	StateBaselined     State = "Baselined"
	StateBelowBaseline State = "Below Baseline"
)

// Migration represents a discovered migration file.
type Migration struct {
	Version     string
	Description string
	Type        Type
	Script      string // filename
	Checksum    uint32
	SQL         string // raw content
}

// AppliedMigration represents a row in drift_schema_history.
type AppliedMigration struct {
	InstalledRank int
	Version       string
	Description   string
	Type          Type
	Script        string
	Checksum      uint32
	InstalledBy   string
	InstalledOn   time.Time
	ExecutionTime int // milliseconds
	Success       bool
}

// CallbackEvent names for lifecycle hooks.
const (
	BeforeMigrate  = "beforeMigrate"
	AfterMigrate   = "afterMigrate"
	BeforeEach     = "beforeEachMigrate"
	AfterEach      = "afterEachMigrate"
	BeforeUndo     = "beforeUndo"
	AfterUndo      = "afterUndo"
	BeforeClean    = "beforeClean"
	AfterClean     = "afterClean"
	BeforeValidate = "beforeValidate"
	AfterValidate  = "afterValidate"
	BeforeRepair   = "beforeRepair"
	AfterRepair    = "afterRepair"
	BeforeBaseline = "beforeBaseline"
	AfterBaseline  = "afterBaseline"
)

var (
	// V001__create_users.sql
	versionedRe = regexp.MustCompile(`^V(\d+)__(.+)\.sql$`)
	// R__refresh_views.sql
	repeatableRe = regexp.MustCompile(`^R__(.+)\.sql$`)
	// U001__drop_users.sql
	undoRe = regexp.MustCompile(`^U(\d+)__(.+)\.sql$`)
	// beforeMigrate__audit.sql
	callbackRe = regexp.MustCompile(`^(before|after)(Migrate|EachMigrate|Undo|Clean|Validate|Repair|Baseline)(?:__(.+))?\.sql$`)
)

// ParseFilename extracts migration metadata from a filename.
func ParseFilename(filename string) (*Migration, error) {
	if m := versionedRe.FindStringSubmatch(filename); m != nil {
		return &Migration{
			Version:     m[1],
			Description: cleanDescription(m[2]),
			Type:        TypeVersioned,
			Script:      filename,
		}, nil
	}

	if m := repeatableRe.FindStringSubmatch(filename); m != nil {
		return &Migration{
			Description: cleanDescription(m[1]),
			Type:        TypeRepeatable,
			Script:      filename,
		}, nil
	}

	if m := undoRe.FindStringSubmatch(filename); m != nil {
		return &Migration{
			Version:     m[1],
			Description: cleanDescription(m[2]),
			Type:        TypeUndo,
			Script:      filename,
		}, nil
	}

	if m := callbackRe.FindStringSubmatch(filename); m != nil {
		event := strings.ToLower(m[1][:1]) + m[1][1:] + m[2]
		desc := event
		if m[3] != "" {
			desc = event + ": " + cleanDescription(m[3])
		}
		return &Migration{
			Description: desc,
			Type:        TypeCallback,
			Script:      filename,
		}, nil
	}

	return nil, fmt.Errorf("unrecognized migration filename: %s", filename)
}

func cleanDescription(s string) string {
	return strings.ReplaceAll(s, "_", " ")
}

// SortKey returns a comparable key for ordering versioned migrations.
func (m *Migration) SortKey() string {
	switch m.Type {
	case TypeVersioned, TypeUndo:
		return fmt.Sprintf("V%010s", m.Version)
	case TypeRepeatable:
		return "R_" + m.Description
	case TypeCallback:
		return "C_" + m.Description
	default:
		return m.Script
	}
}
