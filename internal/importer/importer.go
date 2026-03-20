package importer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Config controls import behaviour.
type Config struct {
	DeleteSourceAfterImport bool
	ConflictAction          string // "skip" | "overwrite" | "rename"
	MovieFolderFormat       string // e.g. "{Movie.Title} ({Movie.Year})"
	MovieFileFormat         string // e.g. "{Movie.Title} ({Movie.Year}) {Quality.Full}"
}

// ImportResult describes the outcome of a single file import.
type ImportResult struct {
	SourcePath string
	DestPath   string
	MovieID    int64
	Action     string // "moved" | "skipped" | "renamed" | "error"
	Error      string
}

// Movie holds the metadata needed to format destination paths.
type Movie struct {
	ID     int64
	Title  string
	Year   int
	ImdbID string
}

// Importer handles moving/renaming files into the library.
type Importer struct {
	cfg    *Config
	rootFS string
}

// New creates an Importer.
func New(cfg *Config, rootFolder string) *Importer {
	return &Importer{cfg: cfg, rootFS: rootFolder}
}

// ImportFile moves/renames a single file to the library.
func (im *Importer) ImportFile(_ context.Context, src string, movie Movie) (ImportResult, error) {
	ext := filepath.Ext(src)
	quality := "Unknown"

	folderName := FormatMoviePath(im.cfg.MovieFolderFormat, movie.Title, movie.Year, movie.ImdbID, quality)
	fileName := FormatMoviePath(im.cfg.MovieFileFormat, movie.Title, movie.Year, movie.ImdbID, quality) + ext

	destDir := filepath.Join(im.rootFS, folderName)
	dest := filepath.Join(destDir, fileName)

	result := ImportResult{
		SourcePath: src,
		DestPath:   dest,
		MovieID:    movie.ID,
	}

	_, statErr := os.Stat(dest)
	destExists := statErr == nil

	if destExists {
		switch im.cfg.ConflictAction {
		case "skip":
			result.Action = "skipped"
			return result, nil
		case "rename":
			dest = strings.TrimSuffix(dest, ext) + ".1" + ext
			result.DestPath = dest
			result.Action = "renamed"
		case "overwrite":
			result.Action = "moved"
		default:
			result.Action = "skipped"
			return result, nil
		}
	} else {
		result.Action = "moved"
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		result.Action = "error"
		result.Error = err.Error()
		return result, fmt.Errorf("create dest dir: %w", err)
	}

	if err := moveFile(src, dest); err != nil {
		result.Action = "error"
		result.Error = err.Error()
		return result, fmt.Errorf("move file: %w", err)
	}

	if im.cfg.DeleteSourceAfterImport {
		_ = os.Remove(src)
	}

	return result, nil
}

// CleanEmptyDirs removes empty directories under rootFolder (deepest first).
func (im *Importer) CleanEmptyDirs(_ context.Context) error {
	var dirs []string
	if err := filepath.WalkDir(im.rootFS, func(path string, d os.DirEntry, err error) error {
		if err != nil || path == im.rootFS || !d.IsDir() {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	}); err != nil {
		return err
	}

	for i := len(dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(dirs[i])
		if err != nil {
			continue
		}
		if len(entries) == 0 {
			_ = os.Remove(dirs[i])
		}
	}
	return nil
}

// FormatMoviePath substitutes template tokens with movie metadata.
// Supported tokens: {Movie.Title}, {Movie.Year}, {Movie.ImdbId}, {Quality.Full}
func FormatMoviePath(template, title string, year int, imdbID, quality string) string {
	r := strings.NewReplacer(
		"{Movie.Title}", sanitizeName(title),
		"{Movie.Year}", fmt.Sprintf("%d", year),
		"{Movie.ImdbId}", sanitizeName(imdbID),
		"{Quality.Full}", sanitizeName(quality),
	)
	return r.Replace(template)
}

// sanitizeName removes characters that are invalid in cross-platform file names.
func sanitizeName(s string) string {
	return strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-",
		"*", "", "?", "", "\"", "",
		"<", "", ">", "", "|", "",
	).Replace(s)
}

// moveFile renames src to dst, falling back to a copy+delete for cross-device moves.
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(dst)
		return err
	}
	return os.Remove(src)
}
