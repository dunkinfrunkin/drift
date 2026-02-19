package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newInfoCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show migration status",
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

			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(infos)
			case "yaml":
				return yaml.NewEncoder(os.Stdout).Encode(infos)
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Version", "Description", "State", "Script", "Installed On", "Exec Time"})
				table.SetBorder(false)
				for _, info := range infos {
					installedOn := ""
					if !info.InstalledOn.IsZero() {
						installedOn = info.InstalledOn.Format("2006-01-02 15:04:05")
					}
					execTime := ""
					if info.ExecTime > 0 {
						execTime = fmt.Sprintf("%dms", info.ExecTime.Milliseconds())
					}
					table.Append([]string{
						info.Version,
						info.Description,
						string(info.State),
						info.Script,
						installedOn,
						execTime,
					})
				}
				table.Render()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "output format (table, json, yaml)")

	return cmd
}
