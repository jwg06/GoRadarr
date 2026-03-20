package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
	"github.com/jwg06/goradarr/internal/events"
	"github.com/jwg06/goradarr/internal/filesystem"
)

type RefreshSummary struct {
	FilesScanned int `json:"filesScanned"`
	Matches      int `json:"matches"`
}

func NewLibraryRefreshTask(db *database.DB, cfg *config.Config, logger *slog.Logger) Task {
	interval := time.Duration(cfg.Scheduler.LibraryRefreshMinutes) * time.Minute
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	return Task{
		Name:     "library-refresh",
		Interval: interval,
		Timeout:  2 * time.Minute,
		Run: func(ctx context.Context) error {
			summary, err := RefreshLibrary(ctx, db, cfg)
			if err != nil {
				return err
			}
			if logger != nil {
				logger.Info("library refresh complete", "files", summary.FilesScanned, "matches", summary.Matches)
			}
			events.PublishDefault(events.Event{Type: events.EventDownloadImported, Data: summary})
			return nil
		},
	}
}

func NewHeartbeatTask(cfg *config.Config) Task {
	interval := time.Duration(cfg.Scheduler.HeartbeatSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return Task{
		Name:     "heartbeat",
		Interval: interval,
		Timeout:  interval,
		Run: func(ctx context.Context) error {
			events.PublishDefault(events.Event{Type: events.EventHealthChanged, Data: map[string]any{"status": "ok", "at": time.Now().Format(time.RFC3339)}})
			return nil
		},
	}
}

func RefreshLibrary(ctx context.Context, db *database.DB, cfg *config.Config) (RefreshSummary, error) {
	libraryMovies, rootPaths, err := loadLibraryMovies(ctx, db, cfg.Data.RootDir)
	if err != nil {
		return RefreshSummary{}, err
	}

	var allFiles []filesystem.ScannedFile
	for _, root := range rootPaths {
		files, err := filesystem.ScanRoot(ctx, root)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, files...)
	}

	matches := filesystem.MatchFilesToMovies(allFiles, libraryMovies)
	if err := persistMatches(ctx, db, libraryMovies, matches); err != nil {
		return RefreshSummary{}, err
	}

	return RefreshSummary{FilesScanned: len(allFiles), Matches: len(matches)}, nil
}

func loadLibraryMovies(ctx context.Context, db *database.DB, defaultRoot string) ([]filesystem.LibraryMovie, []string, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, title, year, COALESCE(root_folder_path,''), COALESCE(path,'') FROM movies`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	roots := map[string]struct{}{}
	if defaultRoot != "" {
		roots[defaultRoot] = struct{}{}
	}

	var movies []filesystem.LibraryMovie
	for rows.Next() {
		var movie filesystem.LibraryMovie
		var rootFolder, path string
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Year, &rootFolder, &path); err != nil {
			return nil, nil, err
		}
		movies = append(movies, movie)
		if rootFolder != "" {
			roots[rootFolder] = struct{}{}
		} else if path != "" {
			roots[filepath.Dir(path)] = struct{}{}
		}
	}

	rootPaths := make([]string, 0, len(roots))
	for root := range roots {
		if root != "" {
			rootPaths = append(rootPaths, root)
		}
	}
	return movies, rootPaths, nil
}

func persistMatches(ctx context.Context, db *database.DB, movies []filesystem.LibraryMovie, matches []filesystem.Match) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	matched := map[int64]filesystem.Match{}
	for _, match := range matches {
		if _, exists := matched[match.MovieID]; !exists {
			matched[match.MovieID] = match
		}
	}

	for _, movie := range movies {
		match, ok := matched[movie.ID]
		if !ok {
			if _, err := tx.ExecContext(ctx, `UPDATE movies SET has_file=FALSE, updated_at=CURRENT_TIMESTAMP WHERE id=?`, movie.ID); err != nil {
				return err
			}
			continue
		}

		movieDir := filepath.Dir(match.File.Path)
		if _, err := tx.ExecContext(ctx, `
            UPDATE movies SET has_file=TRUE, path=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, movieDir, movie.ID); err != nil {
			return err
		}

		relativePath := filepath.ToSlash(match.File.RelativePath)
		if strings.HasPrefix(relativePath, "../") {
			relativePath = filepath.Base(match.File.Path)
		}

		if err := upsertMovieFile(ctx, tx, movie.ID, relativePath, match.File.Size, match.File.Path); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func upsertMovieFile(ctx context.Context, tx *sql.Tx, movieID int64, relativePath string, size int64, originalPath string) error {
	_, err := tx.ExecContext(ctx, `
        INSERT INTO movie_files (movie_id, relative_path, size, original_file_path)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(movie_id, relative_path) DO UPDATE SET
            size=excluded.size,
            original_file_path=excluded.original_file_path,
            updated_at=CURRENT_TIMESTAMP`, movieID, relativePath, size, originalPath)
	if err == nil {
		return nil
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM movie_files WHERE movie_id=? AND relative_path=?`, movieID, relativePath)
	if err != nil {
		return fmt.Errorf("delete old movie file: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
        INSERT INTO movie_files (movie_id, relative_path, size, original_file_path)
        VALUES (?, ?, ?, ?)`, movieID, relativePath, size, originalPath)
	if err != nil {
		return fmt.Errorf("insert movie file: %w", err)
	}
	return nil
}
