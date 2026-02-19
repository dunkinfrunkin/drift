package ui

import "embed"

// dist holds the embedded frontend build output.
// When no frontend build exists, we serve a placeholder.
//
//go:embed all:static
var staticFS embed.FS
