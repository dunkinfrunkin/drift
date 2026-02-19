package main

import (
	// Register database drivers via blank imports
	_ "github.com/frankchan/drift/internal/database/mysql"
	_ "github.com/frankchan/drift/internal/database/postgres"
	_ "github.com/frankchan/drift/internal/database/sqlite"

	"github.com/frankchan/drift/internal/cli"
)

func main() {
	cli.Execute()
}
