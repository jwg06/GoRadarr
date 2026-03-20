package movies

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
	"github.com/jwg06/goradarr/internal/events"
	"github.com/jwg06/goradarr/internal/metadata"
)

const posterBaseURL = "https://image.tmdb.org/t/p/w500"

// Movie represents a movie in the library.
type Movie struct {
	ID                  int64      `json:"id"`
	Title               string     `json:"title"`
	SortTitle           string     `json:"sortTitle"`
	TmdbID              int        `json:"tmdbId"`
	ImdbID              string     `json:"imdbId,omitempty"`
	Overview            string     `json:"overview,omitempty"`
	Status              string     `json:"status"`
	InCinemas           *time.Time `json:"inCinemas,omitempty"`
	PhysicalRelease     *time.Time `json:"physicalRelease,omitempty"`
	DigitalRelease      *time.Time `json:"digitalRelease,omitempty"`
	Year                int        `json:"year"`
	Runtime             int        `json:"runtime"`
	Studio              string     `json:"studio,omitempty"`
	CollectionTitle     string     `json:"collectionTitle,omitempty"`
	CollectionTmdbID    int        `json:"collectionTmdbId,omitempty"`
	QualityProfileID    int        `json:"qualityProfileId"`
	RootFolderPath      string     `json:"rootFolderPath,omitempty"`
	Path                string     `json:"path,omitempty"`
	Monitored           bool       `json:"monitored"`
	MinimumAvailability string     `json:"minimumAvailability"`
	HasFile             bool       `json:"hasFile"`
	Added               time.Time  `json:"added"`
	RemotePoster        string     `json:"remotePoster,omitempty"`
	Genres              []string   `json:"genres,omitempty"`
	Ratings             *Ratings   `json:"ratings,omitempty"`
}

type Ratings struct {
	Value float64 `json:"value"`
	Votes int     `json:"votes"`
}

type handler struct {
	db   *database.DB
	tmdb *metadata.Client
}

func RegisterRoutes(r chi.Router, db *database.DB, cfg *config.Config) {
	h := &handler{db: db}
	if cfg.Metadata.TMDBAPIKey != "" {
		h.tmdb = metadata.NewClient(cfg.Metadata.TMDBAPIKey)
	}
	r.Route("/movie", func(r chi.Router) {
		r.Get("/", h.listMovies)
		r.Post("/", h.addMovie)
		r.Get("/{id}", h.getMovie)
		r.Put("/{id}", h.updateMovie)
		r.Delete("/{id}", h.deleteMovie)
		r.Get("/lookup", h.lookupMovie)
		r.Post("/{id}/command", h.movieCommand)
	})
}

func (h *handler) listMovies(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
        SELECT id, title, COALESCE(sort_title,''), tmdb_id, COALESCE(imdb_id,''),
               COALESCE(overview,''), status, in_cinemas, physical_release, digital_release,
               year, runtime, COALESCE(studio,''), quality_profile_id,
               COALESCE(root_folder_path,''), COALESCE(path,''), monitored,
               minimum_availability, has_file, added
        FROM movies ORDER BY sort_title`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var result []Movie
	for rows.Next() {
		var m Movie
		if err := rows.Scan(
			&m.ID, &m.Title, &m.SortTitle, &m.TmdbID, &m.ImdbID,
			&m.Overview, &m.Status, &m.InCinemas, &m.PhysicalRelease, &m.DigitalRelease,
			&m.Year, &m.Runtime, &m.Studio, &m.QualityProfileID, &m.RootFolderPath,
			&m.Path, &m.Monitored, &m.MinimumAvailability, &m.HasFile, &m.Added,
		); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		result = append(result, m)
	}
	if result == nil {
		result = []Movie{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) getMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var m Movie
	err = h.db.QueryRowContext(r.Context(), `
        SELECT id, title, COALESCE(sort_title,''), tmdb_id, COALESCE(imdb_id,''),
               COALESCE(overview,''), status, in_cinemas, physical_release, digital_release,
               year, runtime, COALESCE(studio,''), quality_profile_id,
               COALESCE(root_folder_path,''), COALESCE(path,''), monitored,
               minimum_availability, has_file, added
        FROM movies WHERE id = ?`, id).Scan(
		&m.ID, &m.Title, &m.SortTitle, &m.TmdbID, &m.ImdbID,
		&m.Overview, &m.Status, &m.InCinemas, &m.PhysicalRelease, &m.DigitalRelease,
		&m.Year, &m.Runtime, &m.Studio, &m.QualityProfileID, &m.RootFolderPath,
		&m.Path, &m.Monitored, &m.MinimumAvailability, &m.HasFile, &m.Added,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "movie not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *handler) addMovie(w http.ResponseWriter, r *http.Request) {
	var m Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if m.Title == "" || m.TmdbID == 0 {
		writeError(w, http.StatusBadRequest, "title and tmdbId are required")
		return
	}
	if m.SortTitle == "" {
		m.SortTitle = m.Title
	}
	if m.Status == "" {
		m.Status = "announced"
	}
	if m.MinimumAvailability == "" {
		m.MinimumAvailability = "released"
	}
	if m.QualityProfileID == 0 {
		m.QualityProfileID = 1
	}

	res, err := h.db.ExecContext(r.Context(), `
        INSERT INTO movies (title, sort_title, tmdb_id, imdb_id, overview, status,
            in_cinemas, physical_release, digital_release, year, runtime, studio,
            quality_profile_id, root_folder_path, path, monitored, minimum_availability)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Title, m.SortTitle, m.TmdbID, m.ImdbID, m.Overview, m.Status,
		m.InCinemas, m.PhysicalRelease, m.DigitalRelease, m.Year, m.Runtime, m.Studio,
		m.QualityProfileID, m.RootFolderPath, m.Path, m.Monitored, m.MinimumAvailability,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	m.ID, _ = res.LastInsertId()
	m.HasFile = false
	m.Added = time.Now()
	events.PublishDefault(events.Event{Type: events.EventMovieAdded, Data: map[string]any{"id": m.ID, "title": m.Title}})
	writeJSON(w, http.StatusCreated, m)
}

func (h *handler) updateMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var m Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	m.ID = id

	_, err = h.db.ExecContext(r.Context(), `
        UPDATE movies SET title=?, sort_title=?, imdb_id=?, overview=?, status=?,
            in_cinemas=?, physical_release=?, digital_release=?, year=?, runtime=?, studio=?,
            quality_profile_id=?, root_folder_path=?, path=?, monitored=?,
            minimum_availability=?, updated_at=CURRENT_TIMESTAMP
        WHERE id=?`,
		m.Title, m.SortTitle, m.ImdbID, m.Overview, m.Status,
		m.InCinemas, m.PhysicalRelease, m.DigitalRelease, m.Year, m.Runtime, m.Studio,
		m.QualityProfileID, m.RootFolderPath, m.Path, m.Monitored, m.MinimumAvailability, m.ID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	events.PublishDefault(events.Event{Type: events.EventMovieUpdated, Data: map[string]any{"id": m.ID, "title": m.Title}})
	writeJSON(w, http.StatusOK, m)
}

func (h *handler) deleteMovie(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	_, err = h.db.ExecContext(r.Context(), "DELETE FROM movies WHERE id = ?", id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	events.PublishDefault(events.Event{Type: events.EventMovieDeleted, Data: map[string]any{"id": id}})
	w.WriteHeader(http.StatusOK)
}

func (h *handler) lookupMovie(w http.ResponseWriter, r *http.Request) {
	term := strings.TrimSpace(r.URL.Query().Get("term"))
	if term == "" {
		writeError(w, http.StatusBadRequest, "term query parameter required")
		return
	}
	if h.tmdb == nil {
		writeError(w, http.StatusServiceUnavailable, "tmdb api key not configured")
		return
	}

	results, err := h.tmdb.SearchMovies(r.Context(), term, 0)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	movies := make([]Movie, 0, len(results))
	for _, item := range results {
		movie := Movie{
			Title:        item.Title,
			SortTitle:    item.Title,
			TmdbID:       item.ID,
			Overview:     item.Overview,
			Monitored:    true,
			Status:       "announced",
			HasFile:      false,
			RemotePoster: posterURL(item.PosterPath),
		}
		if item.ReleaseDate != "" {
			if releaseDate, err := time.Parse("2006-01-02", item.ReleaseDate); err == nil {
				movie.Year = releaseDate.Year()
				movie.InCinemas = &releaseDate
				movie.DigitalRelease = &releaseDate
			}
		}
		if item.VoteAverage > 0 {
			movie.Ratings = &Ratings{Value: item.VoteAverage}
		}
		movies = append(movies, movie)
	}
	writeJSON(w, http.StatusOK, movies)
}

func (h *handler) movieCommand(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	events.PublishDefault(events.Event{Type: events.EventTaskStarted, Data: map[string]any{"command": "movie.command", "movieId": id}})
	w.WriteHeader(http.StatusAccepted)
}

func posterURL(path string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf("%s%s", posterBaseURL, path)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}
