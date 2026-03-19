package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func runMigrations(db *DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	files, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, f := range files {
		version := strings.TrimSuffix(strings.TrimPrefix(f, "migrations/"), ".sql")

		var exists bool
		row := db.QueryRow("SELECT COUNT(1) FROM schema_migrations WHERE version = ?", version)
		if err := row.Scan(&exists); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if exists {
			continue
		}

		data, err := migrationsFS.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(data)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}
	return nil
}

func queryRow[T any](db *sql.DB, scan func(*sql.Row) (T, error), q string, args ...any) (T, error) {
	return scan(db.QueryRow(q, args...))
}
