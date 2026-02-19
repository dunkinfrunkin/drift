package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/frankchan/drift/internal/diff"
)

func newSnapshotCmd() *cobra.Command {
	var (
		output string
		format string
	)

	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Capture the current database schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			_, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			snap, err := diff.CaptureSchema(ctx, db)
			if err != nil {
				return fmt.Errorf("capturing schema: %w", err)
			}

			w := os.Stdout
			if output != "" {
				f, err := os.Create(output)
				if err != nil {
					return err
				}
				defer f.Close()
				w = f
			}

			switch format {
			case "yaml":
				return yaml.NewEncoder(w).Encode(snap)
			default:
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(snap)
			}
		},
	}

	cmd.Flags().StringVar(&output, "output", "", "write to file instead of stdout")
	cmd.Flags().StringVar(&format, "format", "json", "output format (json, yaml)")

	return cmd
}
