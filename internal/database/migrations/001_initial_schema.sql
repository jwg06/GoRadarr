-- 001_initial_schema.sql
-- Core tables for GoRadarr

CREATE TABLE IF NOT EXISTS root_folders (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    path        TEXT NOT NULL UNIQUE,
    free_space  INTEGER,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS quality_definitions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    quality_id    INTEGER NOT NULL UNIQUE,
    title         TEXT NOT NULL,
    min_size      REAL,
    max_size      REAL,
    preferred_size REAL
);

CREATE TABLE IF NOT EXISTS quality_profiles (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    upgrade_allowed BOOLEAN DEFAULT TRUE,
    cutoff      INTEGER NOT NULL,
    items       TEXT NOT NULL, -- JSON array
    min_format_score INTEGER DEFAULT 0,
    cutoff_format_score INTEGER DEFAULT 0,
    format_items TEXT DEFAULT '[]',
    language    TEXT DEFAULT 'any',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tags (
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    label TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS movies (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    title               TEXT NOT NULL,
    sort_title          TEXT,
    tmdb_id             INTEGER NOT NULL UNIQUE,
    imdb_id             TEXT,
    overview            TEXT,
    status              TEXT DEFAULT 'announced',
    in_cinemas          DATETIME,
    physical_release    DATETIME,
    digital_release     DATETIME,
    year                INTEGER,
    runtime             INTEGER,
    studio              TEXT,
    collection_title    TEXT,
    collection_tmdb_id  INTEGER,
    quality_profile_id  INTEGER REFERENCES quality_profiles(id),
    root_folder_path    TEXT,
    path                TEXT,
    monitored           BOOLEAN DEFAULT TRUE,
    minimum_availability TEXT DEFAULT 'released',
    has_file            BOOLEAN DEFAULT FALSE,
    added               DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_info_sync      DATETIME,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_movies_tmdb_id ON movies(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_movies_monitored ON movies(monitored);
CREATE INDEX IF NOT EXISTS idx_movies_title ON movies(sort_title);

CREATE TABLE IF NOT EXISTS movie_files (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    movie_id            INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    relative_path       TEXT NOT NULL,
    size                INTEGER,
    date_added          DATETIME DEFAULT CURRENT_TIMESTAMP,
    scene_name          TEXT,
    quality             TEXT, -- JSON
    media_info          TEXT, -- JSON
    original_file_path  TEXT,
    release_group       TEXT,
    edition             TEXT,
    languages           TEXT, -- JSON
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_movie_files_movie_id ON movie_files(movie_id);

CREATE TABLE IF NOT EXISTS movie_tags (
    movie_id INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    tag_id   INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (movie_id, tag_id)
);

CREATE TABLE IF NOT EXISTS indexers (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT NOT NULL UNIQUE,
    implementation  TEXT NOT NULL,
    config_contract TEXT NOT NULL,
    settings        TEXT DEFAULT '{}', -- JSON
    enable_rss      BOOLEAN DEFAULT TRUE,
    enable_automatic_search BOOLEAN DEFAULT TRUE,
    enable_interactive_search BOOLEAN DEFAULT TRUE,
    priority        INTEGER DEFAULT 25,
    download_client_id INTEGER DEFAULT 0,
    tags            TEXT DEFAULT '[]', -- JSON
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS download_clients (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT NOT NULL UNIQUE,
    implementation  TEXT NOT NULL,
    config_contract TEXT NOT NULL,
    settings        TEXT DEFAULT '{}', -- JSON
    enable          BOOLEAN DEFAULT TRUE,
    priority        INTEGER DEFAULT 1,
    remove_completed_downloads BOOLEAN DEFAULT TRUE,
    remove_failed_downloads    BOOLEAN DEFAULT TRUE,
    tags            TEXT DEFAULT '[]', -- JSON
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notification_configs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT NOT NULL UNIQUE,
    implementation  TEXT NOT NULL,
    config_contract TEXT NOT NULL,
    settings        TEXT DEFAULT '{}', -- JSON
    on_grab         BOOLEAN DEFAULT FALSE,
    on_download     BOOLEAN DEFAULT FALSE,
    on_upgrade      BOOLEAN DEFAULT FALSE,
    on_rename       BOOLEAN DEFAULT FALSE,
    on_health_issue BOOLEAN DEFAULT FALSE,
    on_delete       BOOLEAN DEFAULT FALSE,
    tags            TEXT DEFAULT '[]', -- JSON
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS history (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    movie_id        INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    source_title    TEXT NOT NULL,
    quality         TEXT, -- JSON
    event_type      TEXT NOT NULL,
    date            DATETIME DEFAULT CURRENT_TIMESTAMP,
    download_id     TEXT,
    data            TEXT DEFAULT '{}', -- JSON extra data
    languages       TEXT DEFAULT '[]'  -- JSON
);

CREATE INDEX IF NOT EXISTS idx_history_movie_id ON history(movie_id);
CREATE INDEX IF NOT EXISTS idx_history_date ON history(date);
CREATE INDEX IF NOT EXISTS idx_history_event_type ON history(event_type);

CREATE TABLE IF NOT EXISTS queue_items (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    movie_id            INTEGER REFERENCES movies(id) ON DELETE SET NULL,
    title               TEXT NOT NULL,
    size                INTEGER,
    size_left           INTEGER,
    time_left           TEXT,
    estimated_completion_time DATETIME,
    status              TEXT,
    tracked_download_status TEXT,
    tracked_download_state  TEXT,
    status_messages     TEXT DEFAULT '[]', -- JSON
    download_id         TEXT,
    protocol            TEXT,
    download_client     TEXT,
    output_path         TEXT,
    quality             TEXT, -- JSON
    languages           TEXT DEFAULT '[]', -- JSON
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT OR IGNORE INTO config (key, value) VALUES
    ('first_run', 'true'),
    ('analytics_enabled', 'true'),
    ('update_mechanism', 'built_in'),
    ('update_automatically', 'false'),
    ('proxy_enabled', 'false'),
    ('log_level', 'info'),
    ('cleanup_level', 'none'),
    ('recycle_bin', ''),
    ('certificate_validation', 'enabled'),
    ('authentication_method', 'none'),
    ('launch_browser', 'true');

INSERT OR IGNORE INTO quality_profiles (name, cutoff, items, upgrade_allowed) VALUES
    ('Any', 4, '[{"quality":{"id":0,"name":"Unknown"},"allowed":true},{"quality":{"id":24,"name":"WORKPRINT"},"allowed":true},{"quality":{"id":25,"name":"CAM"},"allowed":true},{"quality":{"id":26,"name":"TELESYNC"},"allowed":true},{"quality":{"id":27,"name":"TELECINE"},"allowed":true},{"quality":{"id":29,"name":"REGIONAL"},"allowed":true},{"quality":{"id":28,"name":"DVDSCR"},"allowed":true},{"quality":{"id":1,"name":"SDTV"},"allowed":true},{"quality":{"id":2,"name":"DVD"},"allowed":true},{"quality":{"id":23,"name":"DVD-R"},"allowed":true},{"quality":{"id":8,"name":"WEBDL-480p"},"allowed":true},{"quality":{"id":12,"name":"WEBRip-480p"},"allowed":true},{"quality":{"id":20,"name":"Bluray-480p"},"allowed":true},{"quality":{"id":21,"name":"Bluray-576p"},"allowed":true},{"quality":{"id":4,"name":"HDTV-720p"},"allowed":true},{"quality":{"id":5,"name":"WEBDL-720p"},"allowed":true},{"quality":{"id":14,"name":"WEBRip-720p"},"allowed":true},{"quality":{"id":6,"name":"Bluray-720p"},"allowed":true},{"quality":{"id":9,"name":"HDTV-1080p"},"allowed":true},{"quality":{"id":3,"name":"WEBDL-1080p"},"allowed":true},{"quality":{"id":15,"name":"WEBRip-1080p"},"allowed":true},{"quality":{"id":7,"name":"Bluray-1080p"},"allowed":true},{"quality":{"id":30,"name":"Remux-1080p"},"allowed":true},{"quality":{"id":16,"name":"HDTV-2160p"},"allowed":true},{"quality":{"id":18,"name":"WEBDL-2160p"},"allowed":true},{"quality":{"id":17,"name":"WEBRip-2160p"},"allowed":true},{"quality":{"id":19,"name":"Bluray-2160p"},"allowed":true},{"quality":{"id":31,"name":"Remux-2160p"},"allowed":true},{"quality":{"id":22,"name":"BR-DISK"},"allowed":true}]', TRUE),
    ('HD-1080p', 7, '[{"quality":{"id":3,"name":"WEBDL-1080p"},"allowed":true},{"quality":{"id":15,"name":"WEBRip-1080p"},"allowed":true},{"quality":{"id":7,"name":"Bluray-1080p"},"allowed":true},{"quality":{"id":30,"name":"Remux-1080p"},"allowed":true}]', TRUE),
    ('HD-720p/1080p', 6, '[{"quality":{"id":5,"name":"WEBDL-720p"},"allowed":true},{"quality":{"id":14,"name":"WEBRip-720p"},"allowed":true},{"quality":{"id":6,"name":"Bluray-720p"},"allowed":true},{"quality":{"id":3,"name":"WEBDL-1080p"},"allowed":true},{"quality":{"id":15,"name":"WEBRip-1080p"},"allowed":true},{"quality":{"id":7,"name":"Bluray-1080p"},"allowed":true}]', TRUE),
    ('Ultra-HD', 19, '[{"quality":{"id":18,"name":"WEBDL-2160p"},"allowed":true},{"quality":{"id":17,"name":"WEBRip-2160p"},"allowed":true},{"quality":{"id":19,"name":"Bluray-2160p"},"allowed":true},{"quality":{"id":31,"name":"Remux-2160p"},"allowed":true}]', TRUE);
