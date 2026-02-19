package engine

import (
	"context"
	"fmt"
	"io"

	"github.com/frankchan/drift/internal/database"
)

// CallbackRunner executes lifecycle callback SQL files.
type CallbackRunner struct {
	db       database.Database
	resolver *Resolver
	output   io.Writer
}

// NewCallbackRunner creates a new CallbackRunner.
func NewCallbackRunner(db database.Database, resolver *Resolver, output io.Writer) *CallbackRunner {
	return &CallbackRunner{db: db, resolver: resolver, output: output}
}

// Run executes all callbacks for the given event.
func (c *CallbackRunner) Run(ctx context.Context, event string) error {
	callbacks, err := c.resolver.ResolveCallbacks(event)
	if err != nil {
		return fmt.Errorf("resolving callbacks for %s: %w", event, err)
	}

	for _, cb := range callbacks {
		fmt.Fprintf(c.output, "  Executing callback: %s\n", cb.Script)
		for _, stmt := range splitSQL(cb.SQL) {
			if stmt == "" {
				continue
			}
			if _, err := c.db.DB().ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("callback %s: %w", cb.Script, err)
			}
		}
	}

	return nil
}
