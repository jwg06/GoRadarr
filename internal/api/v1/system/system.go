package system

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
	"github.com/jwg06/goradarr/internal/logging"
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
		r.Get("/backup", h.listBackups)
		r.Post("/backup", h.createBackup)
		r.Delete("/backup/{name}", h.deleteBackup)
		r.Post("/logs/test", h.testLogs)
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
		"appName":                "GoRadarr",
		"instanceName":           "GoRadarr",
		"version":                "0.1.0",
		"buildTime":              startTime.Format(time.RFC3339),
		"isDebug":                false,
		"isProduction":           true,
		"isAdmin":                false,
		"isUserInteractive":      false,
		"startupPath":            ".",
		"appData":                h.cfg.Data.RootDir,
		"osName":                 runtime.GOOS,
		"osVersion":              runtime.GOARCH,
		"isNetCore":              false,
		"isLinux":                runtime.GOOS == "linux",
		"isOsx":                  runtime.GOOS == "darwin",
		"isWindows":              runtime.GOOS == "windows",
		"isDocker":               false,
		"mode":                   "console",
		"branch":                 "main",
		"authentication":         "none",
		"sqliteVersion":          "3",
		"migrationVersion":       1,
		"urlBase":                h.cfg.BaseURL,
		"runtimeVersion":         runtime.Version(),
		"runtimeName":            "Go",
		"startTime":              startTime.Format(time.RFC3339),
		"packageVersion":         "",
		"packageAuthor":          "",
		"packageUpdateMechanism": "builtIn",
	})
}

// HealthIssue describes a single health check problem.
type HealthIssue struct {
	Source  string `json:"source"`
	Type    string `json:"type"`
	Message string `json:"message"`
	WikiURL string `json:"wikiUrl"`
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	var issues []HealthIssue

	// Check 1: database reachable
	if err := h.db.PingContext(r.Context()); err != nil {
		issues = append(issues, HealthIssue{
			Source:  "DatabaseMigration",
			Type:    "error",
			Message: "Database not available: " + err.Error(),
		})
	}

	// Check 2: disk space < 1 GB on data dir
	if h.cfg.Data.RootDir != "" {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(h.cfg.Data.RootDir, &stat); err == nil {
			free := int64(stat.Bavail) * int64(stat.Bsize)
			const oneGB = int64(1) << 30
			if free < oneGB {
				issues = append(issues, HealthIssue{
					Source:  "DiskSpace",
					Type:    "warning",
					Message: "Less than 1 GB free on data directory",
				})
			}
		}
	}

	// Check 3: TMDB API key missing
	if h.cfg.Metadata.TMDBAPIKey == "" {
		issues = append(issues, HealthIssue{
			Source:  "MetadataProvider",
			Type:    "info",
			Message: "TMDB API key not configured. Metadata features will be limited.",
		})
	}

	if issues == nil {
		writeJSON(w, http.StatusOK, []HealthIssue{})
		return
	}
	writeJSON(w, http.StatusOK, issues)
}

func (h *handler) diskspace(w http.ResponseWriter, r *http.Request) {
	paths := map[string]struct{}{}
	if h.cfg.Data.RootDir != "" {
		paths[h.cfg.Data.RootDir] = struct{}{}
	}

	rows, err := h.db.QueryContext(r.Context(), `SELECT path FROM root_folders ORDER BY path`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var path string
			if scanErr := rows.Scan(&path); scanErr == nil && path != "" {
				paths[path] = struct{}{}
			}
		}
	}

	result := make([]map[string]any, 0, len(paths))
	for path := range paths {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(path, &stat); err != nil {
			continue
		}
		total := int64(stat.Blocks) * int64(stat.Bsize)
		free := int64(stat.Bavail) * int64(stat.Bsize)
		result = append(result, map[string]any{
			"path":       path,
			"label":      filepath.Base(path),
			"freeSpace":  free,
			"totalSpace": total,
		})
	}

	writeJSON(w, http.StatusOK, result)
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
		"logTarget":             h.cfg.LogTarget,
		"logFile":               h.cfg.LogFile,
		"syslogAddress":         h.cfg.SyslogAddress,
		"syslogPort":            h.cfg.SyslogPort,
		"syslogNetwork":         h.cfg.SyslogNetwork,
		"consoleLogLevel":       "info",
		"instanceName":          "GoRadarr",
		"urlBase":               h.cfg.BaseURL,
		"certificateValidation": "enabled",
		"proxyEnabled":          false,
	})
}

func (h *handler) putHostConfig(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid request body"})
		return
	}

	// Apply each recognised field onto the live config pointer.
	if v, ok := body["bindAddress"].(string); ok {
		h.cfg.Host = v
	}
	if v, ok := body["port"].(float64); ok {
		h.cfg.Port = int(v)
	}
	if v, ok := body["urlBase"].(string); ok {
		h.cfg.BaseURL = v
	}
	if v, ok := body["logLevel"].(string); ok {
		h.cfg.LogLevel = v
	}
	if v, ok := body["logTarget"].(string); ok {
		h.cfg.LogTarget = v
	}
	if v, ok := body["logFile"].(string); ok {
		h.cfg.LogFile = v
	}
	if v, ok := body["syslogAddress"].(string); ok {
		h.cfg.SyslogAddress = v
	}
	if v, ok := body["syslogPort"].(float64); ok {
		h.cfg.SyslogPort = int(v)
	}
	if v, ok := body["syslogNetwork"].(string); ok {
		h.cfg.SyslogNetwork = v
	}

	// Reinitialise the global logger with the new settings.
	slog.SetDefault(logging.Setup(h.cfg.LogLevel, h.cfg.LogTarget, h.cfg.LogFile, logging.SyslogConfig{
		Address: h.cfg.SyslogAddress,
		Port:    h.cfg.SyslogPort,
		Network: h.cfg.SyslogNetwork,
	}))
	slog.Info("logging reconfigured", "target", h.cfg.LogTarget, "level", h.cfg.LogLevel,
		"syslog_address", h.cfg.SyslogAddress, "syslog_port", h.cfg.SyslogPort)

	// Persist changes to config.yaml so they survive a restart.
	if err := config.SaveToFile(h.cfg); err != nil {
		slog.Warn("could not persist config to disk", "error", err)
	}

	h.getHostConfig(w, r)
}

// testLogs fires one message at every log level so you can verify syslog delivery.
func (h *handler) testLogs(w http.ResponseWriter, r *http.Request) {
	slog.Debug("GoRadarr test log: DEBUG", "target", h.cfg.LogTarget, "syslog_address", h.cfg.SyslogAddress)
	slog.Info("GoRadarr test log: INFO", "target", h.cfg.LogTarget, "syslog_address", h.cfg.SyslogAddress)
	slog.Warn("GoRadarr test log: WARN", "target", h.cfg.LogTarget, "syslog_address", h.cfg.SyslogAddress)
	slog.Error("GoRadarr test log: ERROR", "target", h.cfg.LogTarget, "syslog_address", h.cfg.SyslogAddress)
	writeJSON(w, http.StatusOK, map[string]any{
		"message":       "Test log messages sent at all levels",
		"target":        h.cfg.LogTarget,
		"syslogAddress": h.cfg.SyslogAddress,
		"syslogPort":    h.cfg.SyslogPort,
		"syslogNetwork": h.cfg.SyslogNetwork,
		"levels":        []string{"debug", "info", "warn", "error"},
	})
}

func (h *handler) getUIConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"id":                       1,
		"firstDayOfWeek":           0,
		"calendarWeekColumnHeader": "ddd M/D",
		"movieRuntimeFormat":       "hoursMinutes",
		"shortDateFormat":          "MMM Do YYYY",
		"longDateFormat":           "dddd, MMMM D YYYY",
		"timeFormat":               "h(:mm)a",
		"showRelativeDates":        true,
		"enableColorImpairedMode":  false,
		"movieInfoLanguage":        1,
		"uiLanguage":               1,
		"theme":                    "auto",
	})
}

func (h *handler) putUIConfig(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid request body"})
		return
	}
	h.getUIConfig(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
