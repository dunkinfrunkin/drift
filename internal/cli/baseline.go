package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/frankchan/drift/internal/diff"
)

func newBaselineCmd() *cobra.Command {
	var (
		version string
		fromDB  bool
		output  string
	)

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

			if fromDB {
				// Generate SQL from current schema
				snap, err := diff.CaptureSchema(ctx, db)
				if err != nil {
					return fmt.Errorf("capturing schema: %w", err)
				}

				sql := diff.GenerateCreateSQL(snap)

				if output != "" {
					if err := os.WriteFile(output, []byte(sql), 0644); err != nil {
						return err
					}
					fmt.Printf("Baseline SQL written to %s\n", output)
				} else {
					fmt.Println(sql)
				}

				if version == "" {
					version = "001"
				}
				return eng.Baseline(ctx, version)
			}

			if version == "" {
				version = "001"
			}

			return eng.Baseline(ctx, version)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "version to baseline at (default: 001)")
	cmd.Flags().BoolVar(&fromDB, "from-db", false, "generate baseline SQL from current database schema")
	cmd.Flags().StringVar(&output, "output", "", "write baseline SQL to file (used with --from-db)")

	return cmd
}
