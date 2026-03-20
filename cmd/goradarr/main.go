package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jwg06/goradarr/internal/config"
	sched "github.com/jwg06/goradarr/internal/core/scheduler"
	"github.com/jwg06/goradarr/internal/database"
	"github.com/jwg06/goradarr/internal/server"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-healthcheck" {
		cfg, _ := config.Load()
		url := fmt.Sprintf("http://localhost:%d/api/v1/ping", cfg.Port)
		resp, err := http.Get(url) //nolint:gosec
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if cfg.Scheduler.Enabled {
		runner := sched.NewRunner(logger)
		runner.Add(sched.NewHeartbeatTask(cfg))
		runner.Add(sched.NewLibraryRefreshTask(db, cfg, logger))
		runner.Start(ctx)
		defer runner.Wait()
	}

	srv := server.New(cfg, db, logger)
	if err := srv.Run(ctx); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
