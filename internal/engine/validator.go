package engine

import (
	"fmt"

	"github.com/frankchan/drift/internal/migration"
)

// ValidationError represents a single validation issue.
type ValidationError struct {
	Version string
	Message string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("V%s: %s", v.Version, v.Message)
}

// Validator checks migration integrity.
type Validator struct{}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate compares discovered migrations against applied history.
func (v *Validator) Validate(
	discovered []*migration.Migration,
	applied []migration.AppliedMigration,
) []ValidationError {
	var errors []ValidationError

	// Index discovered by version
	discoveredByVersion := make(map[string]*migration.Migration)
	for _, m := range discovered {
		if m.Type == migration.TypeVersioned {
			discoveredByVersion[m.Version] = m
		}
	}

	for _, a := range applied {
		if !a.Success {
			errors = append(errors, ValidationError{
				Version: a.Version,
				Message: "previous migration failed (use drift repair to fix)",
			})
			continue
		}

		if a.Type != migration.TypeVersioned {
			continue
		}

		m, ok := discoveredByVersion[a.Version]
		if !ok {
			errors = append(errors, ValidationError{
				Version: a.Version,
				Message: "applied migration no longer found on filesystem",
			})
			continue
		}

		if m.Checksum != a.Checksum {
			errors = append(errors, ValidationError{
				Version: a.Version,
				Message: fmt.Sprintf("checksum mismatch (applied: %d, local: %d)", a.Checksum, m.Checksum),
			})
		}
	}

	return errors
}
