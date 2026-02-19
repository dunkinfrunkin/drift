package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInterpolateEnv(t *testing.T) {
	os.Setenv("DRIFT_TEST_VAR", "hello")
	defer os.Unsetenv("DRIFT_TEST_VAR")

	tests := []struct {
		input string
		want  string
	}{
		{"${DRIFT_TEST_VAR}", "hello"},
		{"${DRIFT_MISSING_VAR:-fallback}", "fallback"},
		{"${DRIFT_TEST_VAR:-fallback}", "hello"},
		{"prefix_${DRIFT_TEST_VAR}_suffix", "prefix_hello_suffix"},
		{"no vars here", "no vars here"},
	}

	for _, tt := range tests {
		got := interpolateEnv(tt.input)
		if got != tt.want {
			t.Errorf("interpolateEnv(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "drift.yaml")
	os.Setenv("DRIFT_DB_URL", "postgres://localhost/test")
	defer os.Unsetenv("DRIFT_DB_URL")

	content := `url: ${DRIFT_DB_URL}
locations:
  - db/migrations
  - db/seeds
table: my_history
`
	os.WriteFile(cfgPath, []byte(content), 0644)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.URL != "postgres://localhost/test" {
		t.Errorf("URL = %q, want %q", cfg.URL, "postgres://localhost/test")
	}
	if len(cfg.Locations) != 2 {
		t.Errorf("Locations = %v, want 2 entries", cfg.Locations)
	}
	if cfg.Table != "my_history" {
		t.Errorf("Table = %q, want %q", cfg.Table, "my_history")
	}
}

func TestLoadMissing(t *testing.T) {
	cfg, err := Load("/nonexistent/drift.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Table != "drift_schema_history" {
		t.Errorf("expected defaults, got Table = %q", cfg.Table)
	}
}

func TestDetectDriver(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"postgres://localhost/db", "postgres"},
		{"postgresql://localhost/db", "postgres"},
		{"mysql://localhost/db", "mysql"},
		{"sqlite://path.db", "sqlite"},
		{"data.db", "sqlite"},
		{"data.sqlite3", "sqlite"},
		{"unknown://foo", ""},
	}
	for _, tt := range tests {
		got := DetectDriver(tt.url)
		if got != tt.want {
			t.Errorf("DetectDriver(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
