-- 002_queue_release.sql
-- Add download tracking to queue_items and a releases cache table.

ALTER TABLE queue_items ADD COLUMN download_url       TEXT;
ALTER TABLE queue_items ADD COLUMN download_client_id INTEGER;

CREATE TABLE IF NOT EXISTS releases (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    title        TEXT    NOT NULL,
    guid         TEXT    NOT NULL UNIQUE,
    indexer_id   INTEGER DEFAULT 0,
    indexer_name TEXT    DEFAULT '',
    download_url TEXT    NOT NULL DEFAULT '',
    size         INTEGER DEFAULT 0,
    seeders      INTEGER DEFAULT 0,
    peers        INTEGER DEFAULT 0,
    protocol     TEXT    DEFAULT 'usenet',
    tmdb_id      INTEGER DEFAULT 0,
    imdb_id      TEXT    DEFAULT '',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);
