package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:frontend
var frontendFS embed.FS

func spaFS() http.Handler {
	dist, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		// If no embedded frontend, return placeholder
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<!DOCTYPE html><html><body><h1>GoRadarr</h1><p>API: <a href="/api/v1/system/status">/api/v1/system/status</a></p></body></html>`))
		})
	}
	fileServer := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the exact file; fall back to index.html for SPA routing
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		_, err := dist.Open(path)
		if err != nil {
			// File not found — serve index.html for client-side routing
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
