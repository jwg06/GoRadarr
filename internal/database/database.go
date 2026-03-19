package database

import (
	"database/sql"
	"fmt"

	"github.com/jwg06/goradarr/internal/config"
	_ "modernc.org/sqlite"
)

// DB wraps *sql.DB with driver awareness.
type DB struct {
	*sql.DB
	Driver string
}

func Open(cfg config.DatabaseConfig) (*DB, error) {
	switch cfg.Driver {
	case "sqlite", "":
		db, err := sql.Open("sqlite", cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}
		db.SetMaxOpenConns(1) // sqlite is single-writer
		if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;"); err != nil {
			return nil, fmt.Errorf("sqlite pragma: %w", err)
		}
		return &DB{DB: db, Driver: "sqlite"}, nil
	case "postgres":
		db, err := sql.Open("pgx", cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}
		return &DB{DB: db, Driver: "postgres"}, nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

func Migrate(db *DB) error {
	return runMigrations(db)
}
