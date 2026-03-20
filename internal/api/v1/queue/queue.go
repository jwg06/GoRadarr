package queue

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/database"
	"github.com/jwg06/goradarr/internal/downloadclient"
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
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	// 1. Load queue item.
	var downloadURL, protocol string
	var downloadClientID int64
	err := h.db.QueryRowContext(r.Context(), `
		SELECT COALESCE(download_url,''), COALESCE(protocol,''), COALESCE(download_client_id,0)
		FROM queue_items WHERE id=?`, id).Scan(&downloadURL, &protocol, &downloadClientID)
	if err != nil {
		writeError(w, http.StatusNotFound, "queue item not found")
		return
	}
	if downloadURL == "" {
		writeError(w, http.StatusBadRequest, "no download URL for this queue item")
		return
	}

	// 2. Load download client config.
	var implementation string
	var settings json.RawMessage
	err = h.db.QueryRowContext(r.Context(),
		`SELECT implementation, COALESCE(settings,'{}') FROM download_clients WHERE id=?`,
		downloadClientID).Scan(&implementation, &settings)
	if err != nil {
		writeError(w, http.StatusNotFound, "download client not found")
		return
	}

	// 3. Build the live client.
	client, err := downloadclient.Build(implementation, settings)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. Dispatch to the correct protocol.
	if protocol == "torrent" {
		err = client.AddTorrent(r.Context(), downloadURL, "")
	} else {
		err = client.AddNZB(r.Context(), downloadURL, "")
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 5. Mark as downloading.
	h.db.ExecContext(r.Context(),
		`UPDATE queue_items SET status='downloading', updated_at=CURRENT_TIMESTAMP WHERE id=?`, id)

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
