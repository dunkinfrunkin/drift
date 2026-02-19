package migration

import "strings"

// SubstitutePlaceholders replaces ${key} placeholders in SQL with values from the map.
func SubstitutePlaceholders(sql string, placeholders map[string]string) string {
	result := sql
	for k, v := range placeholders {
		result = strings.ReplaceAll(result, "${"+k+"}", v)
	}
	return result
}
