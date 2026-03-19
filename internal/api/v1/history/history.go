package history

import (
"encoding/json"
"fmt"
"net/http"
"strconv"
"time"

"github.com/go-chi/chi/v5"
"github.com/jwg06/goradarr/internal/database"
)

type HistoryRecord struct {
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

type handler struct{ db *database.DB }

func RegisterRoutes(r chi.Router, db *database.DB) {
h := &handler{db: db}
r.Route("/history", func(r chi.Router) {
r.Get("/", h.list)
r.Get("/movie", h.byMovie)
r.Delete("/{id}", h.delete)
r.Post("/{id}/failed", h.markFailed)
})
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
q := r.URL.Query()
page, _ := strconv.Atoi(q.Get("page"))
if page < 1 { page = 1 }
pageSize, _ := strconv.Atoi(q.Get("pageSize"))
if pageSize < 1 { pageSize = 20 }
offset := (page - 1) * pageSize
sortKey := q.Get("sortKey")
if sortKey == "" { sortKey = "date" }
sortDir := q.Get("sortDir")
if sortDir != "asc" { sortDir = "desc" }

allowed := map[string]string{"date": "date", "movieId": "movie_id", "sourceTitle": "source_title"}
col, ok := allowed[sortKey]
if !ok { col = "date" }

rows, err := h.db.QueryContext(r.Context(), fmt.Sprintf(`
SELECT id, movie_id, source_title,
       COALESCE(quality,'{}'), COALESCE(languages,'[]'),
       date, event_type, COALESCE(download_id,''), COALESCE(data,'{}')
FROM history ORDER BY %s %s LIMIT ? OFFSET ?`, col, sortDir), pageSize, offset)
if err != nil { writeError(w, 500, err.Error()); return }
defer rows.Close()

var records []HistoryRecord
for rows.Next() {
var rec HistoryRecord
rows.Scan(&rec.ID, &rec.MovieID, &rec.SourceTitle, &rec.Quality, &rec.Languages,
&rec.Date, &rec.EventType, &rec.DownloadID, &rec.Data)
records = append(records, rec)
}
if records == nil { records = []HistoryRecord{} }

var total int
h.db.QueryRowContext(r.Context(), "SELECT COUNT(1) FROM history").Scan(&total)
writeJSON(w, 200, map[string]any{
"page": page, "pageSize": pageSize, "sortKey": sortKey,
"sortDirection": sortDir, "totalRecords": total, "records": records,
})
}

func (h *handler) byMovie(w http.ResponseWriter, r *http.Request) {
movieID, err := strconv.ParseInt(r.URL.Query().Get("movieId"), 10, 64)
if err != nil { writeError(w, 400, "movieId required"); return }
rows, err := h.db.QueryContext(r.Context(), `
SELECT id, movie_id, source_title, COALESCE(quality,'{}'), COALESCE(languages,'[]'),
       date, event_type, COALESCE(download_id,''), COALESCE(data,'{}')
FROM history WHERE movie_id=? ORDER BY date DESC`, movieID)
if err != nil { writeError(w, 500, err.Error()); return }
defer rows.Close()
var records []HistoryRecord
for rows.Next() {
var rec HistoryRecord
rows.Scan(&rec.ID, &rec.MovieID, &rec.SourceTitle, &rec.Quality, &rec.Languages,
&rec.Date, &rec.EventType, &rec.DownloadID, &rec.Data)
records = append(records, rec)
}
if records == nil { records = []HistoryRecord{} }
writeJSON(w, 200, records)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
h.db.ExecContext(r.Context(), "DELETE FROM history WHERE id=?", id)
w.WriteHeader(200)
}

func (h *handler) markFailed(w http.ResponseWriter, r *http.Request) {
id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
h.db.ExecContext(r.Context(),
`UPDATE history SET event_type='downloadFailed' WHERE id=?`, id)
w.WriteHeader(200)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(status)
json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, msg string) {
writeJSON(w, status, map[string]string{"message": msg})
}
