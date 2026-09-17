package db

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate applies all embedded SQL migrations in filename order. Statements
// are executed individually because the SQLite driver rejects multi-statement
// strings. Migrations are idempotent (CREATE ... IF NOT EXISTS).
func Migrate(ctx context.Context, db *DB) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		sqlBytes, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := db.RunMigrations(string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		log.Printf("applied migration %s", name)
	}
	return nil
}

// splitStatements splits a migration file into individual SQL statements on
// semicolon boundaries, trimming blank lines and comments. Naive by design:
// our migrations contain no string literals with embedded semicolons.
func splitStatements(migrationSQL string) []string {
	var statements []string
	var current strings.Builder

	for _, line := range strings.Split(migrationSQL, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		current.WriteString(line)
		current.WriteString(" ")
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, strings.TrimSpace(current.String()))
			current.Reset()
		}
	}

	if strings.TrimSpace(current.String()) != "" {
		statements = append(statements, strings.TrimSpace(current.String()))
	}
	return statements
}