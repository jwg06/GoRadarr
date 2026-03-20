package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseMovieName(t *testing.T) {
	title, year := ParseMovieName("The.Matrix.1999.1080p.BluRay.x264.mkv")
	if title != "The Matrix" {
		t.Fatalf("expected title %q, got %q", "The Matrix", title)
	}
	if year != 1999 {
		t.Fatalf("expected year 1999, got %d", year)
	}
}

func TestMatchFilesToMovies(t *testing.T) {
	matches := MatchFilesToMovies([]ScannedFile{{Title: "The Matrix", Year: 1999}}, []LibraryMovie{{ID: 42, Title: "The Matrix", Year: 1999}})
	if len(matches) != 1 || matches[0].MovieID != 42 {
		t.Fatalf("unexpected matches: %#v", matches)
	}
}

func TestBuildMoviePath(t *testing.T) {
	got := BuildMoviePath("The Matrix", 1999, ".mkv")
	want := filepath.Join("The Matrix (1999)", "The Matrix (1999).mkv")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestScanRoot(t *testing.T) {
	root := t.TempDir()
	movieDir := filepath.Join(root, "Movie")
	if err := os.MkdirAll(movieDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(movieDir, "Movie.2024.1080p.mkv")
	if err := os.WriteFile(path, []byte("demo"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := ScanRoot(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Title != "Movie" || files[0].Year != 2024 {
		t.Fatalf("unexpected scanned file: %#v", files[0])
	}
}
