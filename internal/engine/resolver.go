package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/frankchan/drift/internal/migration"
)

// Resolver discovers migration files from configured locations.
type Resolver struct {
	locations    []string
	placeholders map[string]string
}

// NewResolver creates a Resolver for the given locations.
func NewResolver(locations []string, placeholders map[string]string) *Resolver {
	return &Resolver{locations: locations, placeholders: placeholders}
}

// Resolve scans all locations and returns discovered migrations sorted by version.
func (r *Resolver) Resolve() ([]*migration.Migration, error) {
	var all []*migration.Migration

	for _, loc := range r.locations {
		entries, err := os.ReadDir(loc)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading migration directory %s: %w", loc, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
				continue
			}

			m, err := migration.ParseFilename(entry.Name())
			if err != nil {
				continue // skip unrecognized files
			}

			content, err := os.ReadFile(filepath.Join(loc, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
			}

			m.SQL = string(content)
			if len(r.placeholders) > 0 {
				m.SQL = migration.SubstitutePlaceholders(m.SQL, r.placeholders)
			}
			m.Checksum = migration.ComputeChecksum(m.SQL)

			all = append(all, m)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].SortKey() < all[j].SortKey()
	})

	return all, nil
}

// ResolveByType filters resolved migrations by type.
func (r *Resolver) ResolveByType(t migration.Type) ([]*migration.Migration, error) {
	all, err := r.Resolve()
	if err != nil {
		return nil, err
	}
	var filtered []*migration.Migration
	for _, m := range all {
		if m.Type == t {
			filtered = append(filtered, m)
		}
	}
	return filtered, nil
}

// ResolveCallbacks returns callback migrations for a given event.
func (r *Resolver) ResolveCallbacks(event string) ([]*migration.Migration, error) {
	all, err := r.ResolveByType(migration.TypeCallback)
	if err != nil {
		return nil, err
	}
	var matched []*migration.Migration
	for _, m := range all {
		if matchesCallbackEvent(m, event) {
			matched = append(matched, m)
		}
	}
	return matched, nil
}

func matchesCallbackEvent(m *migration.Migration, event string) bool {
	// Callback description starts with the event name
	return len(m.Description) >= len(event) && m.Description[:len(event)] == event
}
