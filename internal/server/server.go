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
	"github.com/jwg06/goradarr/internal/api/v1/calendar"
	"github.com/jwg06/goradarr/internal/api/v1/downloadclients"
	"github.com/jwg06/goradarr/internal/api/v1/history"
	"github.com/jwg06/goradarr/internal/api/v1/indexers"
	"github.com/jwg06/goradarr/internal/api/v1/movies"
	"github.com/jwg06/goradarr/internal/api/v1/notifications"
	"github.com/jwg06/goradarr/internal/api/v1/profiles"
	"github.com/jwg06/goradarr/internal/api/v1/queue"
	"github.com/jwg06/goradarr/internal/api/v1/system"
	"github.com/jwg06/goradarr/internal/api/v1/tags"
	"github.com/jwg06/goradarr/internal/auth"
	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
	"github.com/jwg06/goradarr/internal/events"
)

type Server struct {
	cfg    *config.Config
	db     *database.DB
	logger *slog.Logger
	broker *events.Broker
	http   *http.Server
}

func New(cfg *config.Config, db *database.DB, logger *slog.Logger) *Server {
	broker := events.NewBroker(logger)
	events.SetDefaultBroker(broker)
	s := &Server{cfg: cfg, db: db, logger: logger, broker: broker}
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
		s.logger.Info("shutting down...")
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
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Api-Key"},
		MaxAge:         300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.APIKeyMiddleware(s.cfg.Auth.APIKey, s.cfg.Auth.Enabled))
		movies.RegisterRoutes(r, s.db, s.cfg)
		profiles.RegisterRoutes(r, s.db)
		history.RegisterRoutes(r, s.db)
		calendar.RegisterRoutes(r, s.db)
		indexers.RegisterRoutes(r, s.db)
		downloadclients.RegisterRoutes(r, s.db)
		notifications.RegisterRoutes(r, s.db)
		queue.RegisterRoutes(r, s.db)
		tags.RegisterRoutes(r, s.db)
		system.RegisterRoutes(r, s.cfg, s.db)
		r.Get("/feed", s.broker.ServeHTTP)
	})

	r.Handle("/*", spaFS())
	return r
}
