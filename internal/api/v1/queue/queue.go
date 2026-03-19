package queue

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/database"
)

type handler struct{ db *database.DB }

func RegisterRoutes(r chi.Router, db *database.DB) {
	h := &handler{db: db}
	r.Route("/queue", func(r chi.Router) {
		r.Get("/", h.list)
		r.Delete("/", h.deleteMany)
		r.Get("/status", h.status)
		r.Get("/details", h.details)
		r.Delete("/{id}", h.deleteOne)
		r.Put("/{id}/grab", h.grab)
	})
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, COALESCE(movie_id,0), title, COALESCE(size,0), COALESCE(size_left,0),
		       COALESCE(status,''), COALESCE(tracked_download_status,''),
		       COALESCE(tracked_download_state,''), COALESCE(download_id,''),
		       COALESCE(protocol,''), COALESCE(download_client,''),
		       COALESCE(quality,'{}'), COALESCE(languages,'[]')
		FROM queue_items ORDER BY created_at DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type QueueRecord struct {
		ID                   int64           `json:"id"`
		MovieID              int64           `json:"movieId"`
		Title                string          `json:"title"`
		Size                 int64           `json:"size"`
		SizeLeft             int64           `json:"sizeleft"`
		Status               string          `json:"status"`
		TrackedDownloadStatus string         `json:"trackedDownloadStatus"`
		TrackedDownloadState  string         `json:"trackedDownloadState"`
		DownloadID           string          `json:"downloadId"`
		Protocol             string          `json:"protocol"`
		DownloadClient       string          `json:"downloadClient"`
		Quality              json.RawMessage `json:"quality"`
		Languages            json.RawMessage `json:"languages"`
	}

	var records []QueueRecord
	for rows.Next() {
		var rec QueueRecord
		rows.Scan(&rec.ID, &rec.MovieID, &rec.Title, &rec.Size, &rec.SizeLeft,
			&rec.Status, &rec.TrackedDownloadStatus, &rec.TrackedDownloadState,
			&rec.DownloadID, &rec.Protocol, &rec.DownloadClient,
			&rec.Quality, &rec.Languages)
		records = append(records, rec)
	}
	if records == nil {
		records = []QueueRecord{}
	}

	var total int
	h.db.QueryRowContext(r.Context(), "SELECT COUNT(1) FROM queue_items").Scan(&total)

	writeJSON(w, http.StatusOK, map[string]any{
		"page":         page,
		"pageSize":     pageSize,
		"sortKey":      "timeLeft",
		"totalRecords": total,
		"records":      records,
	})
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	var total int
	h.db.QueryRowContext(r.Context(), "SELECT COUNT(1) FROM queue_items").Scan(&total)
	writeJSON(w, http.StatusOK, map[string]any{
		"totalCount":           total,
		"count":                total,
		"unknownCount":         0,
		"errors":               false,
		"warnings":             false,
		"unknownErrors":        false,
		"unknownWarnings":      false,
	})
}

func (h *handler) details(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []any{})
}

func (h *handler) deleteOne(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.db.ExecContext(r.Context(), "DELETE FROM queue_items WHERE id=?", id)
	w.WriteHeader(http.StatusOK)
}

func (h *handler) deleteMany(w http.ResponseWriter, r *http.Request) {
	// bulk delete by IDs query param
	ids := r.URL.Query()["id"]
	for _, sid := range ids {
		if id, err := strconv.ParseInt(sid, 10, 64); err == nil {
			h.db.ExecContext(r.Context(), "DELETE FROM queue_items WHERE id=?", id)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *handler) grab(w http.ResponseWriter, r *http.Request) {
	// TODO: trigger download client grab for this queue item
	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}
