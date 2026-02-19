package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/frankchan/drift/internal/engine"
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
			case "html":
				fmt.Println(generateHTMLReport(infos, applied, pending, failed))
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

func generateHTMLReport(infos []engine.MigrationInfo, applied, pending, failed int) string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>drift Migration Report</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f172a;color:#e2e8f0;padding:2rem}
.container{max-width:900px;margin:0 auto}
h1{font-size:1.5rem;margin-bottom:0.5rem}
.meta{color:#64748b;font-size:0.875rem;margin-bottom:2rem}
.stats{display:grid;grid-template-columns:repeat(4,1fr);gap:1rem;margin-bottom:2rem}
.stat{background:#1e293b;border-radius:0.5rem;padding:1.25rem}
.stat-label{color:#64748b;font-size:0.75rem;text-transform:uppercase;letter-spacing:0.05em}
.stat-value{font-size:1.5rem;font-weight:700;margin-top:0.25rem}
.applied{color:#22c55e}.pending{color:#eab308}.failed{color:#ef4444}
table{width:100%;border-collapse:collapse;background:#1e293b;border-radius:0.5rem;overflow:hidden}
th{text-align:left;padding:0.75rem 1rem;color:#64748b;font-size:0.75rem;text-transform:uppercase;letter-spacing:0.05em;border-bottom:1px solid #334155}
td{padding:0.75rem 1rem;border-bottom:1px solid #334155;font-size:0.875rem}
.badge{display:inline-block;padding:0.125rem 0.5rem;border-radius:9999px;font-size:0.75rem;font-weight:500}
.badge-applied{background:#22c55e20;color:#22c55e}.badge-pending{background:#eab30820;color:#eab308}.badge-failed{background:#ef444420;color:#ef4444}
</style>
</head>
<body>
<div class="container">
<h1>drift Migration Report</h1>
`)
	b.WriteString(fmt.Sprintf(`<p class="meta">Generated: %s</p>`, time.Now().Format(time.RFC3339)))

	b.WriteString(`<div class="stats">`)
	b.WriteString(fmt.Sprintf(`<div class="stat"><div class="stat-label">Total</div><div class="stat-value">%d</div></div>`, len(infos)))
	b.WriteString(fmt.Sprintf(`<div class="stat"><div class="stat-label">Applied</div><div class="stat-value applied">%d</div></div>`, applied))
	b.WriteString(fmt.Sprintf(`<div class="stat"><div class="stat-label">Pending</div><div class="stat-value pending">%d</div></div>`, pending))
	b.WriteString(fmt.Sprintf(`<div class="stat"><div class="stat-label">Failed</div><div class="stat-value failed">%d</div></div>`, failed))
	b.WriteString(`</div>`)

	b.WriteString(`<table><thead><tr><th>Version</th><th>Description</th><th>State</th><th>Script</th><th>Installed On</th></tr></thead><tbody>`)
	for _, info := range infos {
		badgeClass := "badge-" + strings.ToLower(string(info.State))
		installedOn := "-"
		if !info.InstalledOn.IsZero() {
			installedOn = info.InstalledOn.Format("2006-01-02 15:04:05")
		}
		b.WriteString(fmt.Sprintf(`<tr><td>V%s</td><td>%s</td><td><span class="badge %s">%s</span></td><td>%s</td><td>%s</td></tr>`,
			info.Version, info.Description, badgeClass, info.State, info.Script, installedOn))
	}
	b.WriteString(`</tbody></table></div></body></html>`)

	return b.String()
}
