package diff

import (
	"fmt"
	"strings"
)

// FormatReport returns a human-readable report of schema differences.
func FormatReport(d *SchemaDiff) string {
	if len(d.Changes) == 0 {
		return "No changes detected."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Schema Changes (%d):\n", len(d.Changes))
	fmt.Fprintln(&b, strings.Repeat("-", 60))

	for _, c := range d.Changes {
		icon := " "
		switch c.Action {
		case "ADD":
			icon = "+"
		case "DROP":
			icon = "-"
		case "MODIFY":
			icon = "~"
		}
		fmt.Fprintf(&b, "  %s %-6s %-8s %s", icon, c.Action, c.ObjectType, c.Name)
		if c.Details != "" {
			fmt.Fprintf(&b, " (%s)", c.Details)
		}
		fmt.Fprintln(&b)
	}

	return b.String()
}
