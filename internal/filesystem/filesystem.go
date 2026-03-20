package filesystem

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	yearPattern  = regexp.MustCompile(`\b(19|20)\d{2}\b`)
	noisePattern = regexp.MustCompile(`(?i)\b(2160p|1080p|720p|480p|bluray|brrip|bdrip|web[-_. ]dl|webrip|hdtv|remux|x264|x265|h264|h265|hevc|10bit|aac|dts|truehd|atmos|proper|repack|extended|unrated|directors? cut|yify|rarbg)\b`)
	splitPattern = regexp.MustCompile(`[._\-]+`)
	videoExts    = map[string]struct{}{
		".mkv": {}, ".mp4": {}, ".avi": {}, ".mov": {}, ".wmv": {}, ".m4v": {},
	}
)

type ScannedFile struct {
	Path         string
	RelativePath string
	Size         int64
	ModTime      time.Time
	Title        string
	Year         int
	Extension    string
}

type LibraryMovie struct {
	ID    int64
	Title string
	Year  int
}

type Match struct {
	MovieID int64
	File    ScannedFile
}

func ScanRoot(ctx context.Context, root string) ([]ScannedFile, error) {
	files := []ScannedFile{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if _, ok := videoExts[ext]; !ok {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		title, year := ParseMovieName(d.Name())
		files = append(files, ScannedFile{
			Path:         path,
			RelativePath: rel,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			Title:        title,
			Year:         year,
			Extension:    ext,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].RelativePath < files[j].RelativePath })
	return files, nil
}

func ParseMovieName(name string) (string, int) {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	year := 0
	if match := yearPattern.FindString(base); match != "" {
		fmt.Sscanf(match, "%d", &year)
		base = strings.Replace(base, match, "", 1)
	}
	base = noisePattern.ReplaceAllString(base, " ")
	base = splitPattern.ReplaceAllString(base, " ")
	base = strings.Join(strings.Fields(base), " ")
	return strings.TrimSpace(base), year
}

func NormalizeTitle(title string) string {
	title = strings.ToLower(title)
	title = splitPattern.ReplaceAllString(title, " ")
	title = strings.Join(strings.Fields(title), " ")
	return strings.TrimSpace(title)
}

func MatchFilesToMovies(files []ScannedFile, movies []LibraryMovie) []Match {
	if len(files) == 0 || len(movies) == 0 {
		return nil
	}

	byTitleYear := map[string]LibraryMovie{}
	byTitle := map[string]LibraryMovie{}
	for _, movie := range movies {
		normalized := NormalizeTitle(movie.Title)
		if movie.Year > 0 {
			byTitleYear[fmt.Sprintf("%s:%d", normalized, movie.Year)] = movie
		}
		byTitle[normalized] = movie
	}

	matches := []Match{}
	for _, file := range files {
		normalized := NormalizeTitle(file.Title)
		if normalized == "" {
			continue
		}
		if file.Year > 0 {
			if movie, ok := byTitleYear[fmt.Sprintf("%s:%d", normalized, file.Year)]; ok {
				matches = append(matches, Match{MovieID: movie.ID, File: file})
				continue
			}
		}
		if movie, ok := byTitle[normalized]; ok {
			matches = append(matches, Match{MovieID: movie.ID, File: file})
		}
	}
	return matches
}

func BuildMoviePath(title string, year int, ext string) string {
	safeTitle := strings.ReplaceAll(strings.TrimSpace(title), "/", "-")
	safeTitle = strings.Join(strings.Fields(safeTitle), " ")
	folder := safeTitle
	if year > 0 {
		folder = fmt.Sprintf("%s (%d)", safeTitle, year)
	}
	if ext == "" {
		ext = ".mkv"
	}
	filename := fmt.Sprintf("%s%s", folder, ext)
	return filepath.Join(folder, filename)
}
