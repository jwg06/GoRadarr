# Changelog

All notable changes to GoRadarr will be documented in this file.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
Versioning: [Semantic Versioning](https://semver.org/spec/v2.0.0.html)

## [Unreleased]

## [0.1.0] - 2026-03-20

### Added
- Full Go rewrite of Radarr movie collection manager
- REST API with Radarr v3 compatibility
- SQLite + PostgreSQL database support
- TMDB metadata integration
- Quality profiles and custom definitions
- Indexer support (Newznab/Torznab protocol)
- Download client support (qBittorrent, SABnzbd, NZBGet, Deluge)
- Queue management
- History tracking
- Calendar view
- Notification system (Discord, Slack, Webhook, Email)
- Real-time SSE event feed
- React 18 + TypeScript + Tailwind CSS frontend
- Embedded Swagger UI + OpenAPI 3.0 spec
- Multi-arch Docker image (linux/amd64, linux/arm64)
- Background scheduler (library refresh, heartbeat)
- File system scanning and movie matching
- Syslog logging support
- Live download client execution
- Import pipeline with rename/move/conflict handling
- JWT authentication with API key management
- Prometheus-style metrics endpoint (/metrics)
- Backup/restore API
