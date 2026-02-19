package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/frankchan/drift/internal/diff"
)

func newDiffCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "diff [source] [target]",
		Short: "Compare database schemas",
		Long:  "Compare the current database schema against a snapshot or another database.",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			_, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			// Capture current schema
			current, err := diff.CaptureSchema(ctx, db)
			if err != nil {
				return fmt.Errorf("capturing schema: %w", err)
			}

			var target *diff.SchemaSnapshot
			if len(args) > 0 {
				// Load snapshot from file
				target, err = diff.LoadSnapshot(args[0])
				if err != nil {
					return fmt.Errorf("loading snapshot: %w", err)
				}
			} else {
				// Compare against empty schema
				target = &diff.SchemaSnapshot{}
			}

			changes := diff.Compare(target, current)

			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(changes)
			case "sql":
				for _, c := range changes.Changes {
					if c.SQL != "" {
						fmt.Println(c.SQL + ";")
					}
				}
			default:
				if len(changes.Changes) == 0 {
					fmt.Println("No differences found.")
					return nil
				}
				for _, c := range changes.Changes {
					fmt.Printf("  %s %s: %s\n", c.Action, c.ObjectType, c.Name)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "output format (text, json, sql)")

	return cmd
}
