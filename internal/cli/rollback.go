package cli

import (
	"github.com/spf13/cobra"
)

func newRollbackCmd() *cobra.Command {
	var (
		count  int
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "rollback [version]",
		Short: "Roll back applied migrations without undo files",
		Long: `Roll back applied migrations by auto-generating reverse DDL.
Unlike 'drift undo', this does not require U__ undo files.

Works like git reset:
  drift rollback           # roll back the last migration
  drift rollback V003      # roll back everything after V003
  drift rollback --count 3 # roll back the last 3 migrations`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			eng, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			target := ""
			if len(args) > 0 {
				target = args[0]
				count = 0 // target overrides count
			} else if count == 0 {
				count = 1 // default: roll back last one
			}

			_, err = eng.Rollback(ctx, count, target, dryRun)
			return err
		},
	}

	cmd.Flags().IntVar(&count, "count", 0, "number of migrations to roll back (default: 1)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be rolled back without executing")

	return cmd
}
