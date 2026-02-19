package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/frankchan/drift/internal/lint"
)

func newLintCmd() *cobra.Command {
	var rules []string

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint migration files for common issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			linter := lint.NewLinter(rules)
			results, err := linter.LintLocations(cfg.Locations)
			if err != nil {
				return err
			}

			if len(results) == 0 {
				color.Green("All migrations passed linting.")
				return nil
			}

			for _, r := range results {
				severity := color.YellowString("WARN")
				if r.Severity == lint.SeverityError {
					severity = color.RedString("ERROR")
				}
				fmt.Printf("  %s [%s] %s: %s\n", severity, r.Rule, r.File, r.Message)
			}

			hasErrors := false
			for _, r := range results {
				if r.Severity == lint.SeverityError {
					hasErrors = true
					break
				}
			}
			if hasErrors {
				return fmt.Errorf("linting failed with errors")
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&rules, "rules", nil, "specific rules to check")

	return cmd
}
