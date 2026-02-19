package lint

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/frankchan/drift/internal/migration"
)

// Severity levels.
const (
	SeverityWarning = "WARNING"
	SeverityError   = "ERROR"
)

// Result represents a single lint finding.
type Result struct {
	File     string `json:"file"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Line     int    `json:"line,omitempty"`
}

// Rule is a lint check function.
type Rule struct {
	Name    string
	Check   func(filename string, sql string) []Result
}

// Linter runs lint rules against migration files.
type Linter struct {
	enabledRules []string
	rules        []Rule
}

// NewLinter creates a Linter with built-in rules.
func NewLinter(enabledRules []string) *Linter {
	l := &Linter{enabledRules: enabledRules}
	l.rules = builtinRules()
	return l
}

// LintLocations scans migration directories and lints all SQL files.
func (l *Linter) LintLocations(locations []string) ([]Result, error) {
	var allResults []Result

	for _, loc := range locations {
		entries, err := os.ReadDir(loc)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
				continue
			}

			// Verify it's a valid migration file
			if _, err := migration.ParseFilename(entry.Name()); err != nil {
				continue
			}

			content, err := os.ReadFile(filepath.Join(loc, entry.Name()))
			if err != nil {
				return nil, err
			}

			for _, rule := range l.rules {
				if !l.isEnabled(rule.Name) {
					continue
				}
				results := rule.Check(entry.Name(), string(content))
				allResults = append(allResults, results...)
			}
		}
	}

	return allResults, nil
}

func (l *Linter) isEnabled(ruleName string) bool {
	if len(l.enabledRules) == 0 {
		return true // all rules enabled by default
	}
	for _, r := range l.enabledRules {
		if r == ruleName {
			return true
		}
	}
	return false
}

func builtinRules() []Rule {
	return []Rule{
		{Name: "no-drop-table", Check: noDropTable},
		{Name: "no-drop-column", Check: noDropColumn},
		{Name: "no-raw-truncate", Check: noRawTruncate},
		{Name: "require-transaction-safe", Check: requireTransactionSafe},
		{Name: "naming-convention", Check: namingConvention},
		{Name: "no-select-star", Check: noSelectStar},
	}
}

func noDropTable(filename string, sql string) []Result {
	upper := strings.ToUpper(sql)
	if strings.Contains(upper, "DROP TABLE") && !strings.Contains(upper, "DROP TABLE IF EXISTS") {
		return []Result{{
			File:     filename,
			Rule:     "no-drop-table",
			Severity: SeverityError,
			Message:  "DROP TABLE without IF EXISTS is dangerous; use DROP TABLE IF EXISTS",
		}}
	}
	return nil
}

func noDropColumn(filename string, sql string) []Result {
	upper := strings.ToUpper(sql)
	if strings.Contains(upper, "DROP COLUMN") {
		return []Result{{
			File:     filename,
			Rule:     "no-drop-column",
			Severity: SeverityWarning,
			Message:  "DROP COLUMN detected; ensure this is intentional and data has been migrated",
		}}
	}
	return nil
}

func noRawTruncate(filename string, sql string) []Result {
	upper := strings.ToUpper(sql)
	if strings.Contains(upper, "TRUNCATE") {
		return []Result{{
			File:     filename,
			Rule:     "no-raw-truncate",
			Severity: SeverityWarning,
			Message:  "TRUNCATE detected; this is irreversible and bypasses triggers",
		}}
	}
	return nil
}

func requireTransactionSafe(filename string, sql string) []Result {
	upper := strings.ToUpper(sql)
	dangerousOps := []string{"CREATE INDEX CONCURRENTLY", "DROP INDEX CONCURRENTLY", "VACUUM", "REINDEX"}
	for _, op := range dangerousOps {
		if strings.Contains(upper, op) {
			return []Result{{
				File:     filename,
				Rule:     "require-transaction-safe",
				Severity: SeverityWarning,
				Message:  op + " cannot run inside a transaction; ensure migration is configured accordingly",
			}}
		}
	}
	return nil
}

func namingConvention(filename string, _ string) []Result {
	m, err := migration.ParseFilename(filename)
	if err != nil {
		return nil
	}
	if m.Type == migration.TypeVersioned && m.Description == "" {
		return []Result{{
			File:     filename,
			Rule:     "naming-convention",
			Severity: SeverityWarning,
			Message:  "migration has no description in filename",
		}}
	}
	return nil
}

func noSelectStar(filename string, sql string) []Result {
	upper := strings.ToUpper(sql)
	if strings.Contains(upper, "SELECT *") {
		return []Result{{
			File:     filename,
			Rule:     "no-select-star",
			Severity: SeverityWarning,
			Message:  "SELECT * found; prefer explicit column lists in migrations",
		}}
	}
	return nil
}
