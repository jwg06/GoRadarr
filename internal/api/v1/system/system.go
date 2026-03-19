package system

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
)

var startTime = time.Now()

type handler struct {
	cfg *config.Config
	db  *database.DB
}

func RegisterRoutes(r chi.Router, cfg *config.Config, db *database.DB) {
	h := &handler{cfg: cfg, db: db}
	r.Route("/system", func(r chi.Router) {
		r.Get("/status", h.status)
		r.Get("/health", h.health)
		r.Get("/diskspace", h.diskspace)
	})
	r.Route("/config", func(r chi.Router) {
		r.Get("/host", h.getHostConfig)
		r.Put("/host", h.putHostConfig)
		r.Get("/ui", h.getUIConfig)
		r.Put("/ui", h.putUIConfig)
	})
	r.Get("/ping", h.ping)
}

func (h *handler) ping(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "OK"})
}

func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"appName":          "GoRadarr",
		"instanceName":     "GoRadarr",
		"version":          "0.1.0",
		"buildTime":        startTime.Format(time.RFC3339),
		"isDebug":          false,
		"isProduction":     true,
		"isAdmin":          false,
		"isUserInteractive": false,
		"startupPath":      ".",
		"appData":          h.cfg.Data.RootDir,
		"osName":           runtime.GOOS,
		"osVersion":        runtime.GOARCH,
		"isNetCore":        false,
		"isLinux":          runtime.GOOS == "linux",
		"isOsx":            runtime.GOOS == "darwin",
		"isWindows":        runtime.GOOS == "windows",
		"isDocker":         false,
		"mode":             "console",
		"branch":           "main",
		"authentication":   "none",
		"sqliteVersion":    "3",
		"migrationVersion": 1,
		"urlBase":          h.cfg.BaseURL,
		"runtimeVersion":   runtime.Version(),
		"runtimeName":      "Go",
		"startTime":        startTime.Format(time.RFC3339),
		"packageVersion":   "",
		"packageAuthor":    "",
		"packageUpdateMechanism": "builtIn",
	})
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, []map[string]any{})
}

func (h *handler) diskspace(w http.ResponseWriter, r *http.Request) {
	// TODO: implement real disk space queries
	writeJSON(w, http.StatusOK, []map[string]any{})
}

func (h *handler) getHostConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"id":                    1,
		"bindAddress":           h.cfg.Host,
		"port":                  h.cfg.Port,
		"sslPort":               9898,
		"enableSsl":             false,
		"launchBrowser":         true,
		"authenticationMethod":  "none",
		"analyticsEnabled":      false,
		"updateAutomatically":   false,
		"updateMechanism":       "builtIn",
		"branch":                "main",
		"logLevel":              h.cfg.LogLevel,
		"consoleLogLevel":       "info",
		"instanceName":          "GoRadarr",
		"urlBase":               h.cfg.BaseURL,
		"certificateValidation": "enabled",
		"proxyEnabled":          false,
	})
}

func (h *handler) putHostConfig(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	// TODO: persist host config changes
	h.getHostConfig(w, r)
}

func (h *handler) getUIConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"id":                  1,
		"firstDayOfWeek":      0,
		"calendarWeekColumnHeader": "ddd M/D",
		"movieRuntimeFormat":  "hoursMinutes",
		"shortDateFormat":     "MMM Do YYYY",
		"longDateFormat":      "dddd, MMMM D YYYY",
		"timeFormat":          "h(:mm)a",
		"showRelativeDates":   true,
		"enableColorImpairedMode": false,
		"movieInfoLanguage":   1,
		"uiLanguage":          1,
		"theme":               "auto",
	})
}

func (h *handler) putUIConfig(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	// TODO: persist UI config
	h.getUIConfig(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
