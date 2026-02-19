package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate applied migrations against filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			eng, db, err := openEngine(ctx)
			if err != nil {
				return err
			}
			defer db.Close()

			errors, err := eng.Validate(ctx)
			if err != nil {
				return err
			}

			if len(errors) == 0 {
				color.Green("Validation passed. All migrations are valid.")
				return nil
			}

			color.Red("Validation failed with %d error(s):", len(errors))
			for _, e := range errors {
				fmt.Printf("  - %s\n", e)
			}
			return fmt.Errorf("validation failed")
		},
	}
}
