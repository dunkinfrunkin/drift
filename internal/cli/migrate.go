package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/frankchan/drift/internal/engine"
)

func newMigrateCmd() *cobra.Command {
	var (
		dryRun     bool
		target     string
		cherryPick string
		skip       string
		outOfOrder bool
	)

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			eng, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			opts := engine.PlanOptions{
				Target:     target,
				OutOfOrder: outOfOrder,
				DryRun:     dryRun,
			}
			if cherryPick != "" {
				opts.CherryPick = strings.Split(cherryPick, ",")
			}
			if skip != "" {
				opts.Skip = strings.Split(skip, ",")
			}

			_, err = eng.Migrate(ctx, opts)
			return err
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be applied without executing")
	cmd.Flags().StringVar(&target, "target", "", "target version to migrate to")
	cmd.Flags().StringVar(&cherryPick, "cherry-pick", "", "comma-separated versions to cherry-pick")
	cmd.Flags().StringVar(&skip, "skip", "", "comma-separated versions to skip")
	cmd.Flags().BoolVar(&outOfOrder, "out-of-order", false, "allow out-of-order migrations")

	return cmd
}
