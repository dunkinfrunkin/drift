package cli

import (
	"github.com/spf13/cobra"
)

func newRepairCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repair",
		Short: "Repair the schema history table",
		Long:  "Removes failed migration entries and realigns checksums with the filesystem.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			eng, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			return eng.Repair(ctx)
		},
	}
}
