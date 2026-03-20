// Package indexers implements Newznab and Torznab protocol clients
// and the REST API for managing indexer configurations.
package indexers

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/database"
)

// IndexerConfig holds persisted indexer settings.
type IndexerConfig struct {
	ID                      int64           `json:"id"`
	Name                    string          `json:"name"`
	Implementation          string          `json:"implementation"`
	ConfigContract          string          `json:"configContract"`
	Fields                  json.RawMessage `json:"fields"`
	EnableRSS               bool            `json:"enableRss"`
	EnableAutomaticSearch   bool            `json:"enableAutomaticSearch"`
	EnableInteractiveSearch bool            `json:"enableInteractiveSearch"`
	Priority                int             `json:"priority"`
	Tags                    json.RawMessage `json:"tags"`
}

type handler struct{ db *database.DB }

func RegisterRoutes(r chi.Router, db *database.DB) {
	h := &handler{db: db}
	r.Route("/indexer", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
		r.Post("/test", h.testAll)
		r.Post("/{id}/test", h.testOne)
	})
	r.Get("/indexer/schema", h.schema)
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, name, implementation, config_contract, COALESCE(settings,'{}'),
		       enable_rss, enable_automatic_search, enable_interactive_search, priority, COALESCE(tags,'[]')
		FROM indexers ORDER BY name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var result []IndexerConfig
	for rows.Next() {
		var ix IndexerConfig
		rows.Scan(&ix.ID, &ix.Name, &ix.Implementation, &ix.ConfigContract, &ix.Fields,
			&ix.EnableRSS, &ix.EnableAutomaticSearch, &ix.EnableInteractiveSearch, &ix.Priority, &ix.Tags)
		result = append(result, ix)
	}
	if result == nil {
		result = []IndexerConfig{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var ix IndexerConfig
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, name, implementation, config_contract, COALESCE(settings,'{}'),
		       enable_rss, enable_automatic_search, enable_interactive_search, priority, COALESCE(tags,'[]')
		FROM indexers WHERE id=?`, id).Scan(
		&ix.ID, &ix.Name, &ix.Implementation, &ix.ConfigContract, &ix.Fields,
		&ix.EnableRSS, &ix.EnableAutomaticSearch, &ix.EnableInteractiveSearch, &ix.Priority, &ix.Tags)
	if err != nil {
		writeError(w, http.StatusNotFound, "indexer not found")
		return
	}
	writeJSON(w, http.StatusOK, ix)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var ix IndexerConfig
	if err := json.NewDecoder(r.Body).Decode(&ix); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if ix.Priority == 0 {
		ix.Priority = 25
	}
	if ix.Fields == nil {
		ix.Fields = json.RawMessage("{}")
	}
	if ix.Tags == nil {
		ix.Tags = json.RawMessage("[]")
	}
	res, err := h.db.ExecContext(r.Context(), `
		INSERT INTO indexers (name, implementation, config_contract, settings,
		    enable_rss, enable_automatic_search, enable_interactive_search, priority, tags)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		ix.Name, ix.Implementation, ix.ConfigContract, string(ix.Fields),
		ix.EnableRSS, ix.EnableAutomaticSearch, ix.EnableInteractiveSearch, ix.Priority, string(ix.Tags))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ix.ID, _ = res.LastInsertId()
	writeJSON(w, http.StatusCreated, ix)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var ix IndexerConfig
	json.NewDecoder(r.Body).Decode(&ix)
	ix.ID = id
	h.db.ExecContext(r.Context(), `
		UPDATE indexers SET name=?, implementation=?, config_contract=?, settings=?,
		    enable_rss=?, enable_automatic_search=?, enable_interactive_search=?,
		    priority=?, tags=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		ix.Name, ix.Implementation, ix.ConfigContract, string(ix.Fields),
		ix.EnableRSS, ix.EnableAutomaticSearch, ix.EnableInteractiveSearch, ix.Priority, string(ix.Tags), ix.ID)
	writeJSON(w, http.StatusOK, ix)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.db.ExecContext(r.Context(), "DELETE FROM indexers WHERE id=?", id)
	w.WriteHeader(http.StatusOK)
}

func (h *handler) testAll(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []map[string]any{})
}

func (h *handler) testOne(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	var ix IndexerConfig
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, name, implementation, config_contract, COALESCE(settings,'{}'),
		       enable_rss, enable_automatic_search, enable_interactive_search, priority, COALESCE(tags,'[]')
		FROM indexers WHERE id=?`, id).Scan(
		&ix.ID, &ix.Name, &ix.Implementation, &ix.ConfigContract, &ix.Fields,
		&ix.EnableRSS, &ix.EnableAutomaticSearch, &ix.EnableInteractiveSearch, &ix.Priority, &ix.Tags)
	if err != nil {
		writeError(w, http.StatusNotFound, "indexer not found")
		return
	}

	fields := parseIndexerFields(ix.Fields)
	baseURL := indexerFieldString(fields, "baseUrl")
	apiKey := indexerFieldString(fields, "apiKey")

	if baseURL == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"isValid":  false,
			"failures": []map[string]string{{"errorMessage": "indexer URL not configured"}},
		})
		return
	}

	capsURL := fmt.Sprintf("%s/api?t=caps&apikey=%s", baseURL, url.QueryEscape(apiKey))
	req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, capsURL, nil)
	nc := NewNewznabClient(baseURL, apiKey)
	resp, err := nc.client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"isValid":  false,
			"failures": []map[string]string{{"errorMessage": err.Error()}},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		writeJSON(w, http.StatusOK, map[string]any{
			"isValid":  false,
			"failures": []map[string]string{{"errorMessage": fmt.Sprintf("indexer returned HTTP %d", resp.StatusCode)}},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"isValid": true, "failures": []any{}})
}

// parseIndexerFields parses Radarr-style field arrays or plain JSON objects.
func parseIndexerFields(raw json.RawMessage) map[string]json.RawMessage {
	var arr []struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
		result := make(map[string]json.RawMessage, len(arr))
		for _, f := range arr {
			result[f.Name] = f.Value
		}
		return result
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj
	}
	return map[string]json.RawMessage{}
}

func indexerFieldString(fields map[string]json.RawMessage, key string) string {
	v, ok := fields[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		return s
	}
	return strings.Trim(string(v), `"`)
}

func (h *handler) schema(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, IndexerSchemas())
}

// --- Newznab / Torznab protocol client ---

// SearchResult is a normalized result from any indexer.
type SearchResult struct {
	Title       string
	Link        string
	Size        int64
	GUID        string
	PubDate     time.Time
	Indexer     string
	DownloadURL string
	ImdbID      string
	TmdbID      int
	Seeders     int
	Peers       int
	Protocol    string // "usenet" | "torrent"
}

// NewznabClient queries Newznab/Torznab-compatible indexers.
type NewznabClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewNewznabClient(baseURL, apiKey string) *NewznabClient {
	return &NewznabClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type rssRoot    struct{ Channel rssChannel `xml:"channel"` }
type rssChannel struct{ Items []rssItem `xml:"item"` }
type rssItem    struct {
	Title     string     `xml:"title"`
	Link      string     `xml:"link"`
	GUID      string     `xml:"guid"`
	Enclosure *enclosure `xml:"enclosure"`
	Attrs     []nzbAttr  `xml:"http://www.newznab.com/DTD/2010/feeds/attributes/ attr"`
	TorAttrs  []nzbAttr  `xml:"https://torznab.com/schemas/2015/feed attr"`
}
type enclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
}
type nzbAttr struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Search queries the indexer and returns normalized results.
func (c *NewznabClient) Search(ctx context.Context, query string, tmdbID int) ([]SearchResult, error) {
	params := url.Values{
		"t":      {"movie"},
		"apikey": {c.apiKey},
		"q":      {query},
	}
	if tmdbID > 0 {
		params.Set("tmdbid", strconv.Itoa(tmdbID))
	}

	u := fmt.Sprintf("%s/api?%s", c.baseURL, params.Encode())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var root rssRoot
	if err := xml.NewDecoder(resp.Body).Decode(&root); err != nil {
		return nil, fmt.Errorf("newznab parse: %w", err)
	}

	var results []SearchResult
	for _, item := range root.Channel.Items {
		sr := SearchResult{Title: item.Title, GUID: item.GUID, DownloadURL: item.Link}
		if item.Enclosure != nil {
			sr.DownloadURL = item.Enclosure.URL
			sr.Size = item.Enclosure.Length
		}
		for _, a := range append(item.Attrs, item.TorAttrs...) {
			switch a.Name {
			case "size":
				sr.Size, _ = strconv.ParseInt(a.Value, 10, 64)
			case "seeders":
				sr.Seeders, _ = strconv.Atoi(a.Value)
			case "peers":
				sr.Peers, _ = strconv.Atoi(a.Value)
			case "imdbid":
				sr.ImdbID = a.Value
			case "tmdbid":
				sr.TmdbID, _ = strconv.Atoi(a.Value)
			}
		}
		results = append(results, sr)
	}
	return results, nil
}

func IndexerSchemas() []map[string]any {
	return []map[string]any{
		{
			"implementation": "Newznab",
			"configContract": "NewznabSettings",
			"fields": []map[string]any{
				{"name": "baseUrl", "label": "URL", "type": "url"},
				{"name": "apiPath", "label": "API Path", "type": "textbox", "value": "/api"},
				{"name": "apiKey", "label": "API Key", "type": "textbox"},
			},
		},
		{
			"implementation": "Torznab",
			"configContract": "TorznabSettings",
			"fields": []map[string]any{
				{"name": "baseUrl", "label": "URL", "type": "url"},
				{"name": "apiPath", "label": "API Path", "type": "textbox", "value": "/api"},
				{"name": "apiKey", "label": "API Key", "type": "textbox"},
				{"name": "seedCriteria.seedRatio", "label": "Seed Ratio", "type": "textbox"},
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
