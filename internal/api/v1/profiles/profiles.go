package profiles

import (
"encoding/json"
"net/http"
"strconv"

"github.com/go-chi/chi/v5"
"github.com/jwg06/goradarr/internal/database"
)

type QualityProfile struct {
ID                int64           `json:"id"`
Name              string          `json:"name"`
UpgradeAllowed    bool            `json:"upgradeAllowed"`
Cutoff            int             `json:"cutoff"`
Items             json.RawMessage `json:"items"`
MinFormatScore    int             `json:"minFormatScore"`
CutoffFormatScore int             `json:"cutoffFormatScore"`
FormatItems       json.RawMessage `json:"formatItems"`
Language          json.RawMessage `json:"language"`
}

type handler struct{ db *database.DB }

func RegisterRoutes(r chi.Router, db *database.DB) {
h := &handler{db: db}
r.Route("/qualityprofile", func(r chi.Router) {
r.Get("/", h.list)
r.Post("/", h.create)
r.Get("/{id}", h.get)
r.Put("/{id}", h.update)
r.Delete("/{id}", h.delete)
})
r.Route("/qualitydefinition", func(r chi.Router) {
r.Get("/", h.listDefinitions)
r.Put("/{id}", h.updateDefinition)
r.Put("/update", h.bulkUpdateDefinitions)
})
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
rows, err := h.db.QueryContext(r.Context(), `
SELECT id, name, upgrade_allowed, cutoff, items, min_format_score,
       cutoff_format_score, COALESCE(format_items,'[]'), COALESCE(language,'"any"')
FROM quality_profiles ORDER BY name`)
if err != nil { writeError(w, 500, err.Error()); return }
defer rows.Close()
var result []QualityProfile
for rows.Next() {
var p QualityProfile
rows.Scan(&p.ID, &p.Name, &p.UpgradeAllowed, &p.Cutoff, &p.Items,
&p.MinFormatScore, &p.CutoffFormatScore, &p.FormatItems, &p.Language)
result = append(result, p)
}
if result == nil { result = []QualityProfile{} }
writeJSON(w, 200, result)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
var p QualityProfile
err := h.db.QueryRowContext(r.Context(), `
SELECT id, name, upgrade_allowed, cutoff, items, min_format_score,
       cutoff_format_score, COALESCE(format_items,'[]'), COALESCE(language,'"any"')
FROM quality_profiles WHERE id=?`, id).Scan(
&p.ID, &p.Name, &p.UpgradeAllowed, &p.Cutoff, &p.Items,
&p.MinFormatScore, &p.CutoffFormatScore, &p.FormatItems, &p.Language)
if err != nil { writeError(w, 404, "not found"); return }
writeJSON(w, 200, p)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
var p QualityProfile
if err := json.NewDecoder(r.Body).Decode(&p); err != nil { writeError(w, 400, "invalid body"); return }
if p.Items == nil { p.Items = json.RawMessage("[]") }
if p.FormatItems == nil { p.FormatItems = json.RawMessage("[]") }
if p.Language == nil { p.Language = json.RawMessage(`"any"`) }
res, err := h.db.ExecContext(r.Context(), `
INSERT INTO quality_profiles (name, upgrade_allowed, cutoff, items, min_format_score, cutoff_format_score, format_items, language)
VALUES (?,?,?,?,?,?,?,?)`,
p.Name, p.UpgradeAllowed, p.Cutoff, string(p.Items), p.MinFormatScore,
p.CutoffFormatScore, string(p.FormatItems), string(p.Language))
if err != nil { writeError(w, 500, err.Error()); return }
p.ID, _ = res.LastInsertId()
writeJSON(w, 201, p)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
var p QualityProfile
json.NewDecoder(r.Body).Decode(&p)
p.ID = id
h.db.ExecContext(r.Context(), `
UPDATE quality_profiles SET name=?, upgrade_allowed=?, cutoff=?, items=?,
    min_format_score=?, cutoff_format_score=?, format_items=?, language=?
WHERE id=?`,
p.Name, p.UpgradeAllowed, p.Cutoff, string(p.Items),
p.MinFormatScore, p.CutoffFormatScore, string(p.FormatItems), string(p.Language), p.ID)
writeJSON(w, 200, p)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
h.db.ExecContext(r.Context(), "DELETE FROM quality_profiles WHERE id=?", id)
w.WriteHeader(200)
}

func (h *handler) listDefinitions(w http.ResponseWriter, r *http.Request) {
rows, err := h.db.QueryContext(r.Context(), `
SELECT id, quality_id, title, COALESCE(min_size,0), COALESCE(max_size,0), COALESCE(preferred_size,0)
FROM quality_definitions ORDER BY quality_id`)
if err != nil { writeError(w, 500, err.Error()); return }
defer rows.Close()
type QDef struct {
ID            int64          `json:"id"`
Quality       map[string]any `json:"quality"`
Title         string         `json:"title"`
MinSize       float64        `json:"minSize"`
MaxSize       float64        `json:"maxSize"`
PreferredSize float64        `json:"preferredSize"`
}
var result []QDef
for rows.Next() {
var d QDef; var qID int
rows.Scan(&d.ID, &qID, &d.Title, &d.MinSize, &d.MaxSize, &d.PreferredSize)
d.Quality = map[string]any{"id": qID, "name": d.Title}
result = append(result, d)
}
if result == nil { result = []QDef{} }
writeJSON(w, 200, result)
}

func (h *handler) updateDefinition(w http.ResponseWriter, r *http.Request) {
id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
var b struct{ MinSize, MaxSize, PreferredSize float64 `json:"minSize,maxSize,preferredSize"` }
json.NewDecoder(r.Body).Decode(&b)
h.db.ExecContext(r.Context(),
"UPDATE quality_definitions SET min_size=?, max_size=?, preferred_size=? WHERE id=?",
b.MinSize, b.MaxSize, b.PreferredSize, id)
w.WriteHeader(202)
}

func (h *handler) bulkUpdateDefinitions(w http.ResponseWriter, r *http.Request) {
var items []struct {
ID            int64   `json:"id"`
MinSize       float64 `json:"minSize"`
MaxSize       float64 `json:"maxSize"`
PreferredSize float64 `json:"preferredSize"`
}
json.NewDecoder(r.Body).Decode(&items)
for _, item := range items {
h.db.ExecContext(r.Context(),
"UPDATE quality_definitions SET min_size=?, max_size=?, preferred_size=? WHERE id=?",
item.MinSize, item.MaxSize, item.PreferredSize, item.ID)
}
h.listDefinitions(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(status)
json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, msg string) {
writeJSON(w, status, map[string]string{"message": msg})
}
