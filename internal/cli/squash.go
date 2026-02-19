package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/frankchan/drift/internal/migration"
)

func newSquashCmd() *cobra.Command {
	var (
		upTo   string
		output string
	)

	cmd := &cobra.Command{
		Use:   "squash",
		Short: "Squash migrations into a single file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			// Read all versioned migration files
			var allSQL []string
			var lastVersion string

			for _, loc := range cfg.Locations {
				entries, err := os.ReadDir(loc)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return err
				}

				for _, entry := range entries {
					if entry.IsDir() {
						continue
					}
					m, err := migration.ParseFilename(entry.Name())
					if err != nil {
						continue
					}
					if m.Type != migration.TypeVersioned {
						continue
					}
					if upTo != "" && m.Version > upTo {
						continue
					}

					content, err := os.ReadFile(loc + "/" + entry.Name())
					if err != nil {
						return err
					}

					allSQL = append(allSQL, fmt.Sprintf("-- %s", entry.Name()))
					allSQL = append(allSQL, string(content))
					lastVersion = m.Version
				}
			}

			if len(allSQL) == 0 {
				fmt.Println("No migrations to squash.")
				return nil
			}

			squashed := strings.Join(allSQL, "\n\n")

			if output != "" {
				if err := os.WriteFile(output, []byte(squashed), 0644); err != nil {
					return err
				}
				fmt.Printf("Squashed migrations up to V%s into %s\n", lastVersion, output)
			} else {
				fmt.Println(squashed)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&upTo, "up-to", "", "squash migrations up to this version")
	cmd.Flags().StringVar(&output, "output", "", "write to file instead of stdout")

	return cmd
}
