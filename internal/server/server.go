package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jwg06/goradarr/internal/api/v1/activity"
	"github.com/jwg06/goradarr/internal/api/v1/calendar"
	"github.com/jwg06/goradarr/internal/api/v1/command"
	"github.com/jwg06/goradarr/internal/api/v1/downloadclients"
	"github.com/jwg06/goradarr/internal/api/v1/history"
	"github.com/jwg06/goradarr/internal/api/v1/indexers"
	"github.com/jwg06/goradarr/internal/api/v1/movies"
	"github.com/jwg06/goradarr/internal/api/v1/notifications"
	"github.com/jwg06/goradarr/internal/api/v1/profiles"
	"github.com/jwg06/goradarr/internal/api/v1/queue"
	"github.com/jwg06/goradarr/internal/api/v1/release"
	"github.com/jwg06/goradarr/internal/api/v1/system"
	"github.com/jwg06/goradarr/internal/api/v1/tags"
	"github.com/jwg06/goradarr/internal/auth"
	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
	"github.com/jwg06/goradarr/internal/events"
	"github.com/jwg06/goradarr/internal/metrics"
	apimiddleware "github.com/jwg06/goradarr/internal/middleware"
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
	// Start notification dispatcher in the background.
	dispatcher := notifications.NewDispatcher(s.db, s.broker, s.logger)
	go dispatcher.Start(ctx)

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
	// 300 req/min per IP, burst 20 — generous for UI use, protects against abuse.
	rateLimiter := apimiddleware.NewRateLimiter(300, 20)

	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Compress(5))
	r.Use(metrics.RequestMiddleware)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Api-Key", "X-Requested-With"},
		MaxAge:         300,
	}))

	r.Get("/openapi.yaml", openAPISpecHandler())
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})
	r.Handle("/docs/*", http.StripPrefix("/docs/", docsFS()))
	r.Handle("/metrics", metrics.Handler())

	r.Route("/api/v1/auth", func(r chi.Router) {
		auth.RegisterRoutes(r, s.cfg, s.db)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(rateLimiter.Middleware)
		if s.cfg.Auth.Enabled {
			// CSRF protection is only meaningful when auth is on — prevents
			// cross-site requests from hijacking authenticated sessions.
			r.Use(apimiddleware.CSRFGuard)
		}
		r.Use(auth.APIKeyMiddleware(s.cfg.Auth.APIKey, s.cfg.Auth.Enabled))
		movies.RegisterRoutes(r, s.db, s.cfg)
		profiles.RegisterRoutes(r, s.db)
		history.RegisterRoutes(r, s.db)
		calendar.RegisterRoutes(r, s.db)
		indexers.RegisterRoutes(r, s.db)
		downloadclients.RegisterRoutes(r, s.db)
		notifications.RegisterRoutes(r, s.db)
		queue.RegisterRoutes(r, s.db)
		release.RegisterRoutes(r, s.db)
		tags.RegisterRoutes(r, s.db)
		system.RegisterRoutes(r, s.cfg, s.db)
		command.RegisterRoutes(r, s.db, s.cfg)
		activity.RegisterRoutes(r, s.db)
		r.Get("/feed", s.broker.ServeHTTP)
	})

	r.Handle("/*", spaFS())
	return r
}
