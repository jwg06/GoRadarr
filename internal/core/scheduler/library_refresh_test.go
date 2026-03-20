package scheduler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
)

func TestRefreshLibraryMarksMatchedMovies(t *testing.T) {
	root := t.TempDir()
	movieDir := filepath.Join(root, "The Matrix (1999)")
	if err := os.MkdirAll(movieDir, 0o755); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(movieDir, "The.Matrix.1999.1080p.mkv")
	if err := os.WriteFile(filePath, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}

	db, err := database.Open(config.DatabaseConfig{Driver: "sqlite", DSN: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`INSERT INTO movies (title, sort_title, tmdb_id, year, quality_profile_id, monitored, minimum_availability, root_folder_path) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"The Matrix", "The Matrix", 603, 1999, 1, true, "released", root,
	); err != nil {
		t.Fatal(err)
	}

	summary, err := RefreshLibrary(context.Background(), db, &config.Config{Data: config.DataConfig{RootDir: root}})
	if err != nil {
		t.Fatal(err)
	}
	if summary.Matches != 1 {
		t.Fatalf("expected 1 match, got %#v", summary)
	}

	var hasFile bool
	var path string
	if err := db.QueryRow(`SELECT has_file, path FROM movies WHERE tmdb_id = 603`).Scan(&hasFile, &path); err != nil {
		t.Fatal(err)
	}
	if !hasFile {
		t.Fatal("expected movie to be marked as having a file")
	}
	if path != movieDir {
		t.Fatalf("expected path %q, got %q", movieDir, path)
	}
}
