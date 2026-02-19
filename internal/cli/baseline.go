package cli

import (
	"github.com/spf13/cobra"
)

func newBaselineCmd() *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "baseline",
		Short: "Baseline an existing database at a specific version",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			eng, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			if version == "" {
				version = "001"
			}

			return eng.Baseline(ctx, version)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "version to baseline at (default: 001)")

	return cmd
}
