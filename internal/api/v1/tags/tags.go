package tags

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/database"
)

type Tag struct {
	ID    int64  `json:"id"`
	Label string `json:"label"`
}

type handler struct{ db *database.DB }

func RegisterRoutes(r chi.Router, db *database.DB) {
	h := &handler{db: db}
	r.Route("/tag", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	})
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), "SELECT id, label FROM tags ORDER BY label")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var result []Tag
	for rows.Next() {
		var t Tag
		rows.Scan(&t.ID, &t.Label)
		result = append(result, t)
	}
	if result == nil {
		result = []Tag{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var t Tag
	if err := h.db.QueryRowContext(r.Context(), "SELECT id, label FROM tags WHERE id=?", id).Scan(&t.ID, &t.Label); err != nil {
		writeError(w, http.StatusNotFound, "tag not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var t Tag
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := h.db.ExecContext(r.Context(), "INSERT INTO tags (label) VALUES (?)", t.Label)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	t.ID, _ = res.LastInsertId()
	writeJSON(w, http.StatusCreated, t)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var t Tag
	json.NewDecoder(r.Body).Decode(&t)
	t.ID = id
	h.db.ExecContext(r.Context(), "UPDATE tags SET label=? WHERE id=?", t.Label, t.ID)
	writeJSON(w, http.StatusOK, t)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.db.ExecContext(r.Context(), "DELETE FROM tags WHERE id=?", id)
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
