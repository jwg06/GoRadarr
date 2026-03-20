package command

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
)

const bufSize = 50

// CommandRecord represents a dispatched command and its current status.
type CommandRecord struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Status  string    `json:"status"`
	Queued  time.Time `json:"queued"`
	MovieID int64     `json:"movieId,omitempty"`
}

// ring buffer state (package-level so it survives across requests)
var (
	idGen  atomic.Int64
	mu     sync.RWMutex
	ring   [bufSize]*CommandRecord
	head   int // next write slot
	stored int // number of records stored (≤ bufSize)
)

type commandRequest struct {
	Name    string `json:"name"`
	MovieID int64  `json:"movieId,omitempty"`
}

type handler struct {
	db  *database.DB
	cfg *config.Config
}

var validCommands = map[string]bool{
	"RefreshMovie":         true,
	"RenameMovie":         true,
	"RescanMovie":         true,
	"DownloadedMoviesScan": true,
}

// RegisterRoutes mounts command endpoints onto r.
func RegisterRoutes(r chi.Router, db *database.DB, cfg *config.Config) {
	h := &handler{db: db, cfg: cfg}
	r.Get("/command", h.listCommands)
	r.Post("/command", h.createCommand)
	r.Get("/command/{id}", h.getCommand)
}

func (h *handler) createCommand(w http.ResponseWriter, r *http.Request) {
	var req commandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid request body"})
		return
	}
	if !validCommands[req.Name] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "unknown command: " + req.Name})
		return
	}

	rec := &CommandRecord{
		ID:      idGen.Add(1),
		Name:    req.Name,
		Status:  "started",
		Queued:  time.Now().UTC(),
		MovieID: req.MovieID,
	}

	mu.Lock()
	ring[head%bufSize] = rec
	head++
	if stored < bufSize {
		stored++
	}
	mu.Unlock()

	writeJSON(w, http.StatusCreated, rec)
}

func (h *handler) listCommands(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]*CommandRecord, 0, stored)
	for i := 0; i < stored; i++ {
		idx := ((head - stored + i) % bufSize + bufSize) % bufSize
		if ring[idx] != nil {
			result = append(result, ring[idx])
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) getCommand(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid id"})
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	for i := 0; i < stored; i++ {
		idx := ((head - stored + i) % bufSize + bufSize) % bufSize
		if ring[idx] != nil && ring[idx].ID == id {
			writeJSON(w, http.StatusOK, ring[idx])
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"message": "command not found"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
