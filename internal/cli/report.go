package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a migration report",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			eng, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			infos, err := eng.Info(ctx)
			if err != nil {
				return err
			}

			// Count states
			var applied, pending, failed int
			for _, info := range infos {
				switch info.State {
				case "Applied":
					applied++
				case "Pending":
					pending++
				case "Failed":
					failed++
				}
			}

			switch format {
			case "json":
				report := map[string]interface{}{
					"total":   len(infos),
					"applied": applied,
					"pending": pending,
					"failed":  failed,
					"details": infos,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			default:
				fmt.Printf("Migration Report\n")
				fmt.Printf("================\n\n")
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Metric", "Count"})
				table.SetBorder(false)
				table.Append([]string{"Total Migrations", fmt.Sprintf("%d", len(infos))})
				table.Append([]string{"Applied", fmt.Sprintf("%d", applied)})
				table.Append([]string{"Pending", fmt.Sprintf("%d", pending)})
				table.Append([]string{"Failed", fmt.Sprintf("%d", failed)})
				table.Render()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "output format (text, json, html)")

	return cmd
}
