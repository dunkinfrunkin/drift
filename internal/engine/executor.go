package engine

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/frankchan/drift/internal/database"
	"github.com/frankchan/drift/internal/history"
	"github.com/frankchan/drift/internal/migration"
)

// ExecuteResult holds the result of executing a single migration.
type ExecuteResult struct {
	Migration     *migration.Migration
	ExecutionTime time.Duration
	Success       bool
	Error         error
}

// Executor runs migration SQL against the database.
type Executor struct {
	db      database.Database
	history *history.History
	output  io.Writer
}

// NewExecutor creates a new Executor.
func NewExecutor(db database.Database, hist *history.History, output io.Writer) *Executor {
	return &Executor{db: db, history: hist, output: output}
}

// Execute runs a single migration.
func (e *Executor) Execute(ctx context.Context, m *migration.Migration, dryRun bool) *ExecuteResult {
	start := time.Now()

	if dryRun {
		fmt.Fprintf(e.output, "[dry-run] Would execute: %s\n", m.Script)
		for _, stmt := range splitSQL(m.SQL) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			fmt.Fprintf(e.output, "  %s\n", truncate(stmt, 120))
		}
		return &ExecuteResult{
			Migration:     m,
			ExecutionTime: time.Since(start),
			Success:       true,
		}
	}

	err := e.executeSQL(ctx, m)
	elapsed := time.Since(start)

	result := &ExecuteResult{
		Migration:     m,
		ExecutionTime: elapsed,
		Success:       err == nil,
		Error:         err,
	}

	// Record in history (skip for callbacks)
	if m.Type != migration.TypeCallback {
		recordErr := e.history.Record(ctx, m, elapsed, err == nil)
		if recordErr != nil && err == nil {
			result.Error = fmt.Errorf("recording history: %w", recordErr)
			result.Success = false
		}
	}

	return result
}

func (e *Executor) executeSQL(ctx context.Context, m *migration.Migration) error {
	dialect := e.db.Dialect()
	useTransaction := dialect.SupportsTransactionalDDL()

	if useTransaction {
		tx, err := e.db.DB().BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("beginning transaction: %w", err)
		}

		for _, stmt := range splitSQL(m.SQL) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("executing %s: %w", m.Script, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing %s: %w", m.Script, err)
		}
	} else {
		for _, stmt := range splitSQL(m.SQL) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := e.db.DB().ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("executing %s: %w", m.Script, err)
			}
		}
	}

	return nil
}

// splitSQL splits a SQL string into individual statements by semicolons,
// respecting quoted strings and comments.
func splitSQL(sql string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	inLineComment := false
	inBlockComment := false

	runes := []rune(sql)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		// Handle comment transitions
		if !inSingleQuote && !inDoubleQuote {
			if !inBlockComment && ch == '-' && next == '-' {
				inLineComment = true
				current.WriteRune(ch)
				continue
			}
			if inLineComment && ch == '\n' {
				inLineComment = false
				current.WriteRune(ch)
				continue
			}
			if !inLineComment && ch == '/' && next == '*' {
				inBlockComment = true
				current.WriteRune(ch)
				continue
			}
			if inBlockComment && ch == '*' && next == '/' {
				inBlockComment = false
				current.WriteRune(ch)
				current.WriteRune(next)
				i++
				continue
			}
		}

		if inLineComment || inBlockComment {
			current.WriteRune(ch)
			continue
		}

		// Handle quote transitions
		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		} else if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		}

		// Statement separator
		if ch == ';' && !inSingleQuote && !inDoubleQuote {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
			continue
		}

		current.WriteRune(ch)
	}

	// Don't forget the last statement
	if stmt := strings.TrimSpace(current.String()); stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
