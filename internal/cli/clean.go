package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newCleanCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Drop the schema history table",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				fmt.Print("Are you sure you want to clean? This will drop the history table. [y/N] ")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				if !strings.EqualFold(strings.TrimSpace(scanner.Text()), "y") {
					fmt.Println("Aborted.")
					return nil
				}
			}

			ctx := cmd.Context()
			eng, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			return eng.Clean(ctx)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompt")

	return cmd
}
