package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var driver string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new drift project",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create drift.yaml
			if _, err := os.Stat("drift.yaml"); err == nil {
				return fmt.Errorf("drift.yaml already exists")
			}

			var urlExample string
			switch driver {
			case "mysql":
				urlExample = "mysql://user:password@tcp(localhost:3306)/mydb"
			case "sqlite":
				urlExample = "sqlite://drift.db"
			default:
				urlExample = "postgres://user:password@localhost:5432/mydb?sslmode=disable"
			}

			content := fmt.Sprintf(`# drift configuration

# Database connection URL
url: %s

# Directories to scan for migration files (e.g. V001__create_users.sql)
locations:
  - migrations

# Table used by drift to track which migrations have been applied
table: drift_schema_history

# Key-value pairs that will be replaced in migration scripts as ${key} placeholders
# placeholders:
#   schema: public
`, urlExample)

			if err := os.WriteFile("drift.yaml", []byte(content), 0644); err != nil {
				return err
			}

			// Create migrations directory
			if err := os.MkdirAll("migrations", 0755); err != nil {
				return err
			}

			fmt.Println("Initialized drift project:")
			fmt.Println("  - Created drift.yaml")
			fmt.Println("  - Created migrations/")
			fmt.Println()
			fmt.Println("Edit drift.yaml to configure your database URL, then run:")
			fmt.Println("  drift migrate")

			return nil
		},
	}

	cmd.Flags().StringVar(&driver, "driver", "postgres", "database driver (postgres, mysql, sqlite)")

	return cmd
}
