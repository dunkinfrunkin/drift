package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all drift configuration.
type Config struct {
	URL        string   `yaml:"url"`
	Locations  []string `yaml:"locations"`
	Table      string   `yaml:"table"`
	Schemas    []string `yaml:"schemas"`
	Baseline   string   `yaml:"baseline"`
	OutOfOrder bool     `yaml:"outOfOrder"`
	Callbacks  []string `yaml:"callbacks"`
	Placeholders map[string]string `yaml:"placeholders"`
	UI         UIConfig `yaml:"ui"`
}

type UIConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

var envVarRe = regexp.MustCompile(`\$\{([^}]+)\}`)

// Defaults returns a Config with sensible defaults.
func Defaults() *Config {
	return &Config{
		Locations: []string{"migrations"},
		Table:     "drift_schema_history",
		Schemas:   []string{"public"},
		UI: UIConfig{
			Port: 4077,
			Host: "127.0.0.1",
		},
	}
}

// Load reads a drift.yaml file and applies env var interpolation.
func Load(path string) (*Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	expanded := interpolateEnv(string(data))

	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// interpolateEnv replaces ${ENV_VAR} with the environment variable value.
// ${VAR:-default} syntax is supported for defaults.
func interpolateEnv(s string) string {
	return envVarRe.ReplaceAllStringFunc(s, func(match string) string {
		inner := match[2 : len(match)-1] // strip ${ and }
		name, defaultVal, hasDefault := strings.Cut(inner, ":-")
		val := os.Getenv(name)
		if val == "" && hasDefault {
			return defaultVal
		}
		return val
	})
}

// DetectDriver returns the driver name from a DSN URL.
func DetectDriver(url string) string {
	switch {
	case strings.HasPrefix(url, "postgres://"), strings.HasPrefix(url, "postgresql://"):
		return "postgres"
	case strings.HasPrefix(url, "mysql://"):
		return "mysql"
	case strings.HasPrefix(url, "sqlite://"):
		return "sqlite"
	case strings.HasSuffix(url, ".db"), strings.HasSuffix(url, ".sqlite"), strings.HasSuffix(url, ".sqlite3"):
		return "sqlite"
	default:
		return ""
	}
}
