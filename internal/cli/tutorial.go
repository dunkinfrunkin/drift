package cli

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed tutorial/*
var tutorialFS embed.FS

func newTutorialCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tutorial",
		Short: "Interactive tutorial to learn drift",
		Long: `Launch a guided, interactive tutorial that walks you through drift's
core features using a local Postgres database (via Docker).

Covers: migrate, info, validate, snapshot, diff, rollback, dry-run, and lint.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a temp directory and extract tutorial files
			dir, err := os.MkdirTemp("", "drift-tutorial-*")
			if err != nil {
				return fmt.Errorf("creating temp directory: %w", err)
			}
			defer os.RemoveAll(dir)

			// Walk embedded files and write them out
			if err := extractEmbedded(dir, "tutorial"); err != nil {
				return fmt.Errorf("extracting tutorial files: %w", err)
			}

			// Make the script executable
			scriptPath := filepath.Join(dir, "tutorial.sh")
			if err := os.Chmod(scriptPath, 0755); err != nil {
				return fmt.Errorf("setting permissions: %w", err)
			}

			// Run the tutorial script interactively
			c := exec.Command("bash", scriptPath)
			c.Dir = dir
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}

	return cmd
}

func extractEmbedded(destDir, srcDir string) error {
	entries, err := tutorialFS.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			if err := extractEmbedded(destPath, srcPath); err != nil {
				return err
			}
			continue
		}

		data, err := tutorialFS.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}
