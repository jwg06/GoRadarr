package activity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/database"
	"github.com/jwg06/goradarr/internal/logging"
)

type handler struct{ db *database.DB }

func RegisterRoutes(r chi.Router, db *database.DB) {
	h := &handler{db: db}
	r.Route("/log", func(r chi.Router) {
		r.Get("/", h.getLogs)
		r.Get("/stream", h.streamLogs)
	})
	r.Get("/history/recent", h.recentHistory)
}

// getLogs returns the last 200 log entries, optionally filtered by ?level=.
func (h *handler) getLogs(w http.ResponseWriter, r *http.Request) {
	levelFilter := strings.ToUpper(r.URL.Query().Get("level"))
	entries := logging.RecentLogs(200)
	if levelFilter != "" && levelFilter != "ALL" {
		filtered := entries[:0]
		for _, e := range entries {
			if e.Level == levelFilter {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}
	if entries == nil {
		entries = []logging.LogEntry{}
	}
	writeJSON(w, 200, entries)
}

// streamLogs is an SSE endpoint that pushes new log entries in real time.
func (h *handler) streamLogs(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, 500, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := logging.Subscribe()
	defer logging.Unsubscribe(ch)

	enc := json.NewEncoder(w)
	for {
		select {
		case <-r.Context().Done():
			return
		case entry := <-ch:
			fmt.Fprint(w, "data: ")
			enc.Encode(entry) // appends '\n'
			fmt.Fprint(w, "\n")
			flusher.Flush()
		}
	}
}

type historyRecord struct {
	ID          int64           `json:"id"`
	MovieID     int64           `json:"movieId"`
	SourceTitle string          `json:"sourceTitle"`
	Quality     json.RawMessage `json:"quality"`
	Languages   json.RawMessage `json:"languages"`
	Date        time.Time       `json:"date"`
	EventType   string          `json:"eventType"`
	DownloadID  string          `json:"downloadId,omitempty"`
	Data        json.RawMessage `json:"data"`
}

// recentHistory returns the last 20 history records.
func (h *handler) recentHistory(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, movie_id, source_title,
		       COALESCE(quality,'{}'), COALESCE(languages,'[]'),
		       date, event_type, COALESCE(download_id,''), COALESCE(data,'{}')
		FROM history ORDER BY date DESC LIMIT 20`)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var records []historyRecord
	for rows.Next() {
		var rec historyRecord
		rows.Scan(&rec.ID, &rec.MovieID, &rec.SourceTitle, &rec.Quality, &rec.Languages,
			&rec.Date, &rec.EventType, &rec.DownloadID, &rec.Data)
		records = append(records, rec)
	}
	if records == nil {
		records = []historyRecord{}
	}
	writeJSON(w, 200, records)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}
