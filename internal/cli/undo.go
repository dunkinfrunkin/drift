package cli

import (
	"github.com/spf13/cobra"
)

func newUndoCmd() *cobra.Command {
	var (
		count  int
		target string
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Undo applied migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			eng, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			_, err = eng.Undo(ctx, count, target, dryRun)
			return err
		},
	}

	cmd.Flags().IntVar(&count, "count", 0, "number of migrations to undo (default: 1)")
	cmd.Flags().StringVar(&target, "target", "", "undo down to this version (exclusive)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be undone without executing")

	return cmd
}
