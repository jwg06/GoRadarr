package calendar

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/database"
)

type handler struct{ db *database.DB }

func RegisterRoutes(r chi.Router, db *database.DB) {
	h := &handler{db: db}
	r.Get("/calendar", h.calendar)
}

func (h *handler) calendar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	unmonitored := q.Get("unmonitored") == "true"
	now := time.Now()
	start, end := now.AddDate(0, 0, -7), now.AddDate(0, 0, 28)
	if t, err := time.Parse(time.RFC3339, q.Get("start")); err == nil {
		start = t
	} else if t, err := time.Parse("2006-01-02", q.Get("start")); err == nil {
		start = t
	}
	if t, err := time.Parse(time.RFC3339, q.Get("end")); err == nil {
		end = t
	} else if t, err := time.Parse("2006-01-02", q.Get("end")); err == nil {
		end = t
	}

	query := `SELECT id, title, COALESCE(sort_title,''), tmdb_id, COALESCE(imdb_id,''),
		COALESCE(overview,''), status, year, runtime, COALESCE(studio,''),
		quality_profile_id, COALESCE(root_folder_path,''), COALESCE(path,''),
		monitored, minimum_availability, has_file, added,
		in_cinemas, physical_release, digital_release
		FROM movies WHERE (
		(in_cinemas BETWEEN ? AND ?) OR
		(physical_release BETWEEN ? AND ?) OR
		(digital_release BETWEEN ? AND ?)
	)`
	args := []any{start, end, start, end, start, end}
	if !unmonitored {
		query += " AND monitored = TRUE"
	}
	query += " ORDER BY COALESCE(in_cinemas, physical_release, digital_release)"

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer rows.Close()

	type M struct {
		ID              int64      `json:"id"`
		Title           string     `json:"title"`
		SortTitle       string     `json:"sortTitle"`
		TmdbID          int        `json:"tmdbId"`
		ImdbID          string     `json:"imdbId,omitempty"`
		Overview        string     `json:"overview,omitempty"`
		Status          string     `json:"status"`
		Year            int        `json:"year"`
		Runtime         int        `json:"runtime"`
		Studio          string     `json:"studio,omitempty"`
		QualityProfileID int       `json:"qualityProfileId"`
		RootFolderPath  string     `json:"rootFolderPath,omitempty"`
		Path            string     `json:"path,omitempty"`
		Monitored       bool       `json:"monitored"`
		MinimumAvailability string `json:"minimumAvailability"`
		HasFile         bool       `json:"hasFile"`
		Added           time.Time  `json:"added"`
		InCinemas       *time.Time `json:"inCinemas,omitempty"`
		PhysicalRelease *time.Time `json:"physicalRelease,omitempty"`
		DigitalRelease  *time.Time `json:"digitalRelease,omitempty"`
	}

	var result []M
	for rows.Next() {
		var m M
		rows.Scan(&m.ID, &m.Title, &m.SortTitle, &m.TmdbID, &m.ImdbID,
			&m.Overview, &m.Status, &m.Year, &m.Runtime, &m.Studio,
			&m.QualityProfileID, &m.RootFolderPath, &m.Path,
			&m.Monitored, &m.MinimumAvailability, &m.HasFile, &m.Added,
			&m.InCinemas, &m.PhysicalRelease, &m.DigitalRelease)
		result = append(result, m)
	}
	if result == nil {
		result = []M{}
	}
	writeJSON(w, 200, result)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}
