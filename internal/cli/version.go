package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print drift version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("drift %s (commit: %s, built: %s)\n", Version, Commit, Date)
		},
	}
}
