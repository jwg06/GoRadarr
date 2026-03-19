package downloadclients

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/database"
)

// DownloadClientConfig holds persisted download client settings.
type DownloadClientConfig struct {
	ID                       int64           `json:"id"`
	Name                     string          `json:"name"`
	Enable                   bool            `json:"enable"`
	Protocol                 string          `json:"protocol"`
	Priority                 int             `json:"priority"`
	RemoveCompletedDownloads bool            `json:"removeCompletedDownloads"`
	RemoveFailedDownloads    bool            `json:"removeFailedDownloads"`
	Implementation           string          `json:"implementation"`
	ConfigContract           string          `json:"configContract"`
	Fields                   json.RawMessage `json:"fields"`
	Tags                     json.RawMessage `json:"tags"`
}

// Client is the common interface all download clients must satisfy.
type Client interface {
	Name() string
	Protocol() string
	AddTorrent(ctx context.Context, downloadURL string, category string) error
	AddNZB(ctx context.Context, downloadURL string, category string) error
	GetItems(ctx context.Context) ([]QueueItem, error)
	RemoveItem(ctx context.Context, id string, deleteData bool) error
	TestConnection(ctx context.Context) error
}

// QueueItem represents a download in progress.
type QueueItem struct {
	ID            string
	Title         string
	TotalSize     int64
	RemainingSize int64
	TimeLeft      time.Duration
	Status        string
	Category      string
	DownloadID    string
	Protocol      string
}

type handler struct{ db *database.DB }

func RegisterRoutes(r chi.Router, db *database.DB) {
	h := &handler{db: db}
	r.Route("/downloadclient", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
		r.Post("/test", h.testAll)
		r.Post("/{id}/test", h.testOne)
	})
	r.Get("/downloadclient/schema", h.schema)
	r.Route("/remotepathmapping", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, []any{}) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201) })
		r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	})
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, name, enable, priority, remove_completed_downloads, remove_failed_downloads,
		       implementation, config_contract, COALESCE(settings,'{}'), COALESCE(tags,'[]')
		FROM download_clients ORDER BY name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var result []DownloadClientConfig
	for rows.Next() {
		var dc DownloadClientConfig
		rows.Scan(&dc.ID, &dc.Name, &dc.Enable, &dc.Priority,
			&dc.RemoveCompletedDownloads, &dc.RemoveFailedDownloads,
			&dc.Implementation, &dc.ConfigContract, &dc.Fields, &dc.Tags)
		dc.Protocol = protocolFor(dc.Implementation)
		result = append(result, dc)
	}
	if result == nil {
		result = []DownloadClientConfig{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var dc DownloadClientConfig
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, name, enable, priority, remove_completed_downloads, remove_failed_downloads,
		       implementation, config_contract, COALESCE(settings,'{}'), COALESCE(tags,'[]')
		FROM download_clients WHERE id=?`, id).Scan(
		&dc.ID, &dc.Name, &dc.Enable, &dc.Priority,
		&dc.RemoveCompletedDownloads, &dc.RemoveFailedDownloads,
		&dc.Implementation, &dc.ConfigContract, &dc.Fields, &dc.Tags)
	if err != nil {
		writeError(w, http.StatusNotFound, "download client not found")
		return
	}
	dc.Protocol = protocolFor(dc.Implementation)
	writeJSON(w, http.StatusOK, dc)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var dc DownloadClientConfig
	if err := json.NewDecoder(r.Body).Decode(&dc); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if dc.Priority == 0 {
		dc.Priority = 1
	}
	if dc.Fields == nil {
		dc.Fields = json.RawMessage("{}")
	}
	if dc.Tags == nil {
		dc.Tags = json.RawMessage("[]")
	}
	res, err := h.db.ExecContext(r.Context(), `
		INSERT INTO download_clients (name, enable, priority, remove_completed_downloads,
		    remove_failed_downloads, implementation, config_contract, settings, tags)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		dc.Name, dc.Enable, dc.Priority, dc.RemoveCompletedDownloads,
		dc.RemoveFailedDownloads, dc.Implementation, dc.ConfigContract,
		string(dc.Fields), string(dc.Tags))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	dc.ID, _ = res.LastInsertId()
	writeJSON(w, http.StatusCreated, dc)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var dc DownloadClientConfig
	json.NewDecoder(r.Body).Decode(&dc)
	dc.ID = id
	h.db.ExecContext(r.Context(), `
		UPDATE download_clients SET name=?, enable=?, priority=?,
		    remove_completed_downloads=?, remove_failed_downloads=?,
		    implementation=?, config_contract=?, settings=?, tags=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		dc.Name, dc.Enable, dc.Priority, dc.RemoveCompletedDownloads,
		dc.RemoveFailedDownloads, dc.Implementation, dc.ConfigContract,
		string(dc.Fields), string(dc.Tags), dc.ID)
	writeJSON(w, http.StatusOK, dc)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.db.ExecContext(r.Context(), "DELETE FROM download_clients WHERE id=?", id)
	w.WriteHeader(http.StatusOK)
}

func (h *handler) testAll(w http.ResponseWriter, r *http.Request)  { writeJSON(w, 200, []any{}) }
func (h *handler) testOne(w http.ResponseWriter, r *http.Request)  { w.WriteHeader(200) }

func (h *handler) schema(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, ClientSchemas())
}

func protocolFor(impl string) string {
	torrent := map[string]bool{
		"qBittorrent": true, "Deluge": true, "rTorrent": true,
		"Transmission": true, "UTorrent": true,
	}
	if torrent[impl] {
		return "torrent"
	}
	return "usenet"
}

func ClientSchemas() []map[string]any {
	return []map[string]any{
		{
			"implementation": "qBittorrent",
			"configContract": "QBittorrentSettings",
			"protocol":       "torrent",
			"fields": []map[string]any{
				{"name": "host", "label": "Host", "type": "textbox", "value": "localhost"},
				{"name": "port", "label": "Port", "type": "number", "value": 8080},
				{"name": "useSsl", "label": "Use SSL", "type": "checkbox", "value": false},
				{"name": "username", "label": "Username", "type": "textbox"},
				{"name": "password", "label": "Password", "type": "password"},
				{"name": "movieCategory", "label": "Category", "type": "textbox", "value": "radarr"},
			},
		},
		{
			"implementation": "SABnzbd",
			"configContract": "SabnzbdSettings",
			"protocol":       "usenet",
			"fields": []map[string]any{
				{"name": "host", "label": "Host", "type": "textbox", "value": "localhost"},
				{"name": "port", "label": "Port", "type": "number", "value": 8080},
				{"name": "useSsl", "label": "Use SSL", "type": "checkbox", "value": false},
				{"name": "apiKey", "label": "API Key", "type": "textbox"},
				{"name": "movieCategory", "label": "Category", "type": "textbox", "value": "Movies"},
			},
		},
		{
			"implementation": "NzbGet",
			"configContract": "NzbGetSettings",
			"protocol":       "usenet",
			"fields": []map[string]any{
				{"name": "host", "label": "Host", "type": "textbox", "value": "localhost"},
				{"name": "port", "label": "Port", "type": "number", "value": 6789},
				{"name": "username", "label": "Username", "type": "textbox", "value": "nzbget"},
				{"name": "password", "label": "Password", "type": "password"},
				{"name": "movieCategory", "label": "Category", "type": "textbox", "value": "Movies"},
			},
		},
		{
			"implementation": "Deluge",
			"configContract": "DelugeSettings",
			"protocol":       "torrent",
			"fields": []map[string]any{
				{"name": "host", "label": "Host", "type": "textbox", "value": "localhost"},
				{"name": "port", "label": "Port", "type": "number", "value": 8112},
				{"name": "password", "label": "Password", "type": "password"},
				{"name": "movieCategory", "label": "Category", "type": "textbox", "value": "radarr"},
			},
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}
