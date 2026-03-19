package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jwg06/goradarr/internal/api/v1/movies"
	"github.com/jwg06/goradarr/internal/api/v1/system"
	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
)

type Server struct {
	cfg    *config.Config
	db     *database.DB
	logger *slog.Logger
	http   *http.Server
}

func New(cfg *config.Config, db *database.DB, logger *slog.Logger) *Server {
	s := &Server{cfg: cfg, db: db, logger: logger}
	s.http = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      s.buildRouter(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("GoRadarr starting", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("shutting down server...")
		shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return s.http.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

func (s *Server) buildRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Api-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		movies.RegisterRoutes(r, s.db)
		system.RegisterRoutes(r, s.cfg, s.db)
	})

	// Serve frontend SPA
	r.Get("/*", s.spaHandler())

	return r
}

func (s *Server) spaHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<!DOCTYPE html><html><head><title>GoRadarr</title></head><body><h1>GoRadarr</h1><p>Frontend coming soon. API available at <a href='/api/v1'>/api/v1</a></p></body></html>")
	}
}
