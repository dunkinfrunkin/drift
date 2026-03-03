package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/frankchan/drift/internal/config"
	"github.com/frankchan/drift/internal/database"
	"github.com/frankchan/drift/internal/engine"
)

var (
	cfgFile    string
	dbURL      string
	locations  []string
	tableName  string
	schemas    []string
	verbose    bool
	quiet      bool
)

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "drift",
		Short: "Database migration tool",
		Long:  "drift — a fast, open-source database migration tool with undo, dry-run, cherry-pick, drift detection, schema diff, and linting.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "drift.yaml", "config file path")
	root.PersistentFlags().StringVar(&dbURL, "url", "", "database URL")
	root.PersistentFlags().StringSliceVar(&locations, "locations", nil, "migration file locations")
	root.PersistentFlags().StringVar(&tableName, "table", "", "schema history table name")
	root.PersistentFlags().StringSliceVar(&schemas, "schemas", nil, "schemas to manage")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress output")

	root.AddCommand(
		newMigrateCmd(),
		newUndoCmd(),
		newRollbackCmd(),
		newInfoCmd(),
		newValidateCmd(),
		newCleanCmd(),
		newBaselineCmd(),
		newRepairCmd(),
		newDiffCmd(),
		newSnapshotCmd(),
		newReportCmd(),
		newLintCmd(),
		newSquashCmd(),
		newUICmd(),
		newInitCmd(),
		newTutorialCmd(),
		newVersionCmd(),
	)

	return root
}

// Execute runs the root command.
func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}

// loadConfig loads and merges config from file + flags.
func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, err
	}

	// Override with flags
	if dbURL != "" {
		cfg.URL = dbURL
	}
	if len(locations) > 0 {
		cfg.Locations = locations
	}
	if tableName != "" {
		cfg.Table = tableName
	}
	if len(schemas) > 0 {
		cfg.Schemas = schemas
	}

	// Also check environment variable
	if cfg.URL == "" {
		cfg.URL = os.Getenv("DRIFT_URL")
	}

	return cfg, nil
}

// openEngine loads config, opens the database, and returns an Engine.
func openEngine(ctx context.Context) (*engine.Engine, database.Database, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, nil, err
	}

	if cfg.URL == "" {
		return nil, nil, fmt.Errorf("no database URL configured (use --url, drift.yaml, or DRIFT_URL env var)")
	}

	driver := config.DetectDriver(cfg.URL)
	if driver == "" {
		return nil, nil, fmt.Errorf("cannot detect database driver from URL: %s", cfg.URL)
	}

	db, err := database.Open(ctx, driver, cfg.URL)
	if err != nil {
		return nil, nil, err
	}

	eng := engine.New(cfg, db)
	if quiet {
		eng.SetOutput(noopWriter{})
	}

	return eng, db, nil
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }
