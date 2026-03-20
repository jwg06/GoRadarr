package importer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jwg06/goradarr/internal/importer"
)

func TestFormatMoviePath_AllTokens(t *testing.T) {
	tmpl := "{Movie.Title} ({Movie.Year}) [{Movie.ImdbId}] {Quality.Full}"
	got := importer.FormatMoviePath(tmpl, "The Matrix", 1999, "tt0133093", "Bluray-1080p")
	want := "The Matrix (1999) [tt0133093] Bluray-1080p"
	if got != want {
		t.Errorf("FormatMoviePath = %q, want %q", got, want)
	}
}

func TestImportFile_MovesFileToCorrectDestination(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(t.TempDir(), "movie.mkv")
	if err := os.WriteFile(src, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &importer.Config{
		MovieFolderFormat: "{Movie.Title} ({Movie.Year})",
		MovieFileFormat:   "{Movie.Title} ({Movie.Year}) {Quality.Full}",
		ConflictAction:    "skip",
	}
	im := importer.New(cfg, root)
	movie := importer.Movie{ID: 1, Title: "The Matrix", Year: 1999, ImdbID: "tt0133093"}

	res, err := im.ImportFile(context.Background(), src, movie)
	if err != nil {
		t.Fatalf("ImportFile error: %v", err)
	}
	if res.Action != "moved" {
		t.Errorf("expected action=moved, got %q", res.Action)
	}

	expected := filepath.Join(root, "The Matrix (1999)", "The Matrix (1999) Unknown.mkv")
	if res.DestPath != expected {
		t.Errorf("DestPath = %q, want %q", res.DestPath, expected)
	}
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("dest file not found: %v", err)
	}
}

func TestImportFile_ConflictSkip(t *testing.T) {
	root := t.TempDir()
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "movie.mkv")
	if err := os.WriteFile(src, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &importer.Config{
		MovieFolderFormat: "{Movie.Title} ({Movie.Year})",
		MovieFileFormat:   "{Movie.Title} ({Movie.Year}) {Quality.Full}",
		ConflictAction:    "skip",
	}
	im := importer.New(cfg, root)
	movie := importer.Movie{ID: 1, Title: "The Matrix", Year: 1999}

	// Pre-create the destination file
	destDir := filepath.Join(root, "The Matrix (1999)")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	destFile := filepath.Join(destDir, "The Matrix (1999) Unknown.mkv")
	if err := os.WriteFile(destFile, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := im.ImportFile(context.Background(), src, movie)
	if err != nil {
		t.Fatalf("ImportFile error: %v", err)
	}
	if res.Action != "skipped" {
		t.Errorf("expected action=skipped, got %q", res.Action)
	}
	// Source should still exist
	if _, err := os.Stat(src); err != nil {
		t.Errorf("source should still exist after skip: %v", err)
	}
}

func TestImportFile_ConflictOverwrite(t *testing.T) {
	root := t.TempDir()
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "movie.mkv")
	if err := os.WriteFile(src, []byte("new-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &importer.Config{
		MovieFolderFormat: "{Movie.Title} ({Movie.Year})",
		MovieFileFormat:   "{Movie.Title} ({Movie.Year}) {Quality.Full}",
		ConflictAction:    "overwrite",
	}
	im := importer.New(cfg, root)
	movie := importer.Movie{ID: 1, Title: "The Matrix", Year: 1999}

	// Pre-create the destination file
	destDir := filepath.Join(root, "The Matrix (1999)")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	destFile := filepath.Join(destDir, "The Matrix (1999) Unknown.mkv")
	if err := os.WriteFile(destFile, []byte("old-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := im.ImportFile(context.Background(), src, movie)
	if err != nil {
		t.Fatalf("ImportFile error: %v", err)
	}
	if res.Action != "moved" {
		t.Errorf("expected action=moved, got %q", res.Action)
	}

	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(content) != "new-data" {
		t.Errorf("dest content = %q, want %q", string(content), "new-data")
	}
}

func TestCleanEmptyDirs_RemovesNestedEmptyDirs(t *testing.T) {
	root := t.TempDir()

	// Create nested empty dirs: root/a/b/c
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a dir with a file — should NOT be removed
	kept := filepath.Join(root, "kept")
	if err := os.MkdirAll(kept, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(kept, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	im := importer.New(&importer.Config{}, root)
	if err := im.CleanEmptyDirs(context.Background()); err != nil {
		t.Fatalf("CleanEmptyDirs error: %v", err)
	}

	// a/b/c, a/b and a should all be gone
	for _, dir := range []string{nested, filepath.Join(root, "a", "b"), filepath.Join(root, "a")} {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("expected %q to be removed", dir)
		}
	}
	// kept dir should still exist
	if _, err := os.Stat(kept); err != nil {
		t.Errorf("kept dir should still exist: %v", err)
	}
}
