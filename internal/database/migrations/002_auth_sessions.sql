CREATE TABLE IF NOT EXISTS auth_sessions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    username   TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

-- Store generated API key overrides
INSERT OR IGNORE INTO config (key, value) VALUES ('api_key', '');
