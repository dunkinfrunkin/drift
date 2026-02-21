//go:build integration

package integration

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	// Register database drivers
	_ "github.com/frankchan/drift/internal/database/postgres"

	"github.com/frankchan/drift/internal/cli"
)

const (
	pgHost = "127.0.0.1"
	pgPort = "15432"
	pgUser = "drift"
	pgPass = "drift"
	pgDB   = "drift_test"
)

var (
	containerID string
	basePgURL   string

	// test result tracking
	resultsMu sync.Mutex
	results   []testResult
)

type testResult struct {
	Command string
	Passed  bool
	Skipped bool
}

func TestMain(m *testing.M) {
	// Check docker is available and daemon is running
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Println("SKIP: docker not found in PATH")
		os.Exit(0)
	}
	if out, err := exec.Command("docker", "info").CombinedOutput(); err != nil {
		fmt.Printf("SKIP: docker daemon not available: %v\n%s\n", err, out)
		os.Exit(0)
	}

	// Start postgres container
	if err := startPostgres(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start postgres: %v\n", err)
		os.Exit(1)
	}

	basePgURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUser, pgPass, pgHost, pgPort, pgDB)

	// Wait for readiness
	if err := waitForPostgres(30 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "Postgres not ready: %v\n", err)
		stopPostgres()
		os.Exit(1)
	}

	code := m.Run()

	printReport()
	stopPostgres()
	os.Exit(code)
}

func startPostgres() error {
	cmd := exec.Command("docker", "run", "-d",
		"--name", "drift-integration-test",
		"-e", "POSTGRES_USER="+pgUser,
		"-e", "POSTGRES_PASSWORD="+pgPass,
		"-e", "POSTGRES_DB="+pgDB,
		"-p", pgPort+":5432",
		"postgres:16-alpine",
	)
	out, err := cmd.Output()
	if err != nil {
		// Container might already exist from a failed run; remove and retry
		exec.Command("docker", "rm", "-f", "drift-integration-test").Run()
		out, err = exec.Command("docker", "run", "-d",
			"--name", "drift-integration-test",
			"-e", "POSTGRES_USER="+pgUser,
			"-e", "POSTGRES_PASSWORD="+pgPass,
			"-e", "POSTGRES_DB="+pgDB,
			"-p", pgPort+":5432",
			"postgres:16-alpine",
		).Output()
		if err != nil {
			return fmt.Errorf("docker run: %w", err)
		}
	}
	containerID = strings.TrimSpace(string(out))
	return nil
}

func waitForPostgres(timeout time.Duration) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUser, pgPass, pgHost, pgPort, pgDB)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		db, err := sql.Open("pgx", dsn)
		if err == nil {
			if err = db.Ping(); err == nil {
				db.Close()
				return nil
			}
			db.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("postgres not ready after %s", timeout)
}

func stopPostgres() {
	if containerID != "" {
		exec.Command("docker", "rm", "-f", containerID).Run()
	}
}

// setupTestSchema creates a unique PG schema for test isolation.
// Returns a postgres URL with search_path set and a temp migrations dir.
func setupTestSchema(t *testing.T) (pgURL string, migDir string) {
	t.Helper()

	// Sanitize test name into a valid schema name
	schema := strings.ToLower(t.Name())
	schema = strings.ReplaceAll(schema, "/", "_")
	schema = strings.ReplaceAll(schema, " ", "_")
	schema = strings.ReplaceAll(schema, "-", "_")
	// Truncate to 63 chars (PG identifier limit)
	if len(schema) > 63 {
		schema = schema[:63]
	}

	db, err := sql.Open("pgx", basePgURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %q`, schema))
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	t.Cleanup(func() {
		db2, err := sql.Open("pgx", basePgURL)
		if err != nil {
			return
		}
		defer db2.Close()
		db2.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS %q CASCADE`, schema))
	})

	pgURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		pgUser, pgPass, pgHost, pgPort, pgDB, schema)

	migDir = t.TempDir()
	return pgURL, migDir
}

// runCLI executes a drift CLI command in-process, capturing stdout and stderr.
// Must NOT be called from parallel tests (os.Stdout redirect is not goroutine-safe).
func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	// Capture stdout
	origStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// Capture stderr
	origStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// Reset package-level flag vars by creating a fresh command each time
	root := cli.NewRootCmd()
	root.SetArgs(args)

	execErr := root.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout = origStdout
	os.Stderr = origStderr

	var outBuf, errBuf bytes.Buffer
	io.Copy(&outBuf, rOut)
	io.Copy(&errBuf, rErr)

	return outBuf.String(), errBuf.String(), execErr
}

// writeMigration writes a SQL file into the given directory.
func writeMigration(t *testing.T, dir, filename, sqlContent string) {
	t.Helper()
	path := dir + "/" + filename
	if err := os.WriteFile(path, []byte(sqlContent), 0644); err != nil {
		t.Fatalf("write migration %s: %v", filename, err)
	}
}

// writeConfig writes a minimal drift.yaml to the given directory.
func writeConfig(t *testing.T, dir, pgURL, migDir string) string {
	t.Helper()
	cfgPath := dir + "/drift.yaml"
	content := fmt.Sprintf("url: %s\nlocations:\n  - %s\ntable: drift_schema_history\n", pgURL, migDir)
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return cfgPath
}

func recordResult(command string, passed, skipped bool) {
	resultsMu.Lock()
	defer resultsMu.Unlock()
	results = append(results, testResult{Command: command, Passed: passed, Skipped: skipped})
}

func printReport() {
	resultsMu.Lock()
	defer resultsMu.Unlock()

	passed, failed, skipped := 0, 0, 0
	for _, r := range results {
		switch {
		case r.Skipped:
			skipped++
		case r.Passed:
			passed++
		default:
			failed++
		}
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║     Integration Test Summary Report      ║")
	fmt.Println("╠══════════════════════════════════════════╣")
	for _, r := range results {
		status := "PASS"
		if r.Skipped {
			status = "SKIP"
		} else if !r.Passed {
			status = "FAIL"
		}
		fmt.Printf("║  %-6s %-33s ║\n", status, r.Command)
	}
	fmt.Println("╠══════════════════════════════════════════╣")
	fmt.Printf("║  %d passed, %d failed, %d skipped         ║\n", passed, failed, skipped)
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
}
