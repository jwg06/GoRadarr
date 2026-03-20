# GoRadarr

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE)

> **Radarr rebuilt in Go** — blazing-fast movie collection manager with a modern, professional UX.

GoRadarr is a ground-up rewrite of [Radarr](https://github.com/Radarr/Radarr) in Go, targeting significantly lower resource usage, faster startup, single-binary deployment, and a top-tier user experience powered by a React + TypeScript frontend.

## Goals

| Goal | Approach |
|------|----------|
| **Performance** | Go backend, zero GC pauses on hot paths, WAL SQLite |
| **Single binary** | Embeds frontend assets, no external runtime required |
| **Modern UX** | React 18 + TypeScript + Vite + shadcn/ui + Tailwind |
| **API compatible** | Radarr v3 API compatibility for existing apps/plugins |
| **Cross-platform** | Linux, macOS, Windows, Docker |

## Quick Start

```bash
# Build
make build

# Run
./bin/goradarr

# Open http://localhost:7878
```

## Configuration

GoRadarr reads from `~/.config/goradarr/config.yaml` by default. All values are also configurable via `GORADARR_*` environment variables.

```yaml
host: 0.0.0.0
port: 7878
log_level: info

database:
  driver: sqlite   # or "postgres"
  dsn: ~/.config/goradarr/goradarr.db

auth:
  enabled: false
```

## Architecture

```
GoRadarr/
├── cmd/goradarr/          # Entry point
├── internal/
│   ├── api/v1/            # REST handlers (Radarr v3 compatible)
│   │   ├── movies/
│   │   ├── indexers/
│   │   ├── downloadclients/
│   │   ├── notifications/
│   │   ├── queue/
│   │   ├── history/
│   │   ├── calendar/
│   │   ├── profiles/
│   │   └── system/
│   ├── core/              # Business logic
│   │   ├── movies/        # Library management, file matching
│   │   ├── indexers/      # Newznab/Torznab search
│   │   ├── downloadclients/ # qBittorrent, SABnzbd, NZBGet, etc.
│   │   ├── notifications/ # Webhooks, Discord, Slack, email...
│   │   ├── quality/       # Quality profiles & custom formats
│   │   ├── scheduler/     # Background job runner
│   │   └── history/
│   ├── database/          # SQLite + Postgres, migrations
│   ├── config/            # Viper-based config
│   ├── auth/              # API key + basic auth
│   ├── server/            # chi HTTP server
│   ├── events/            # SSE real-time push
│   ├── metadata/          # TMDB client
│   └── filesystem/        # File scanning, renaming
├── frontend/              # React 18 + TypeScript + Vite
│   └── src/
│       ├── components/
│       ├── pages/
│       ├── hooks/
│       ├── store/         # Zustand
│       ├── api/           # TanStack Query
│       └── types/
├── scripts/
├── .github/workflows/
└── Makefile
```

## Radarr Feature Coverage

- [x] Movie library management
- [x] REST API (v3 compatible)
- [x] SQLite + PostgreSQL support
- [x] Quality profiles
- [x] TMDB-backed movie lookup
- [ ] Indexer support (Newznab, Torznab)
- [ ] Download client integrations (qBittorrent, SABnzbd, NZBGet, Deluge, rTorrent)
- [ ] Automatic search & grab
- [x] File management scanning + matching foundation
- [ ] Custom formats
- [x] Calendar view
- [ ] History tracking
- [ ] Notifications (Discord, Slack, email, Webhook)
- [ ] Authentication
- [x] Real-time updates via SSE
- [x] React frontend
- [ ] Auto-updater
- [x] Docker image

## Development

```bash
# Install dependencies
go mod download

# Run in dev mode
go run ./cmd/goradarr

# Run tests
make test

# Build release binary
make release
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22+, chi, modernc/sqlite |
| Frontend | React 18, TypeScript, Vite, Tailwind, shadcn/ui |
| Database | SQLite (default), PostgreSQL |
| Config | Viper |
| Testing | Go standard `testing` + testify |

## License

GPL-3.0 — same as the original Radarr project.
