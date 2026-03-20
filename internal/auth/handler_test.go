package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/auth"
	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
)

func setupTestRouter(t *testing.T) (*chi.Mux, *config.Config) {
	t.Helper()
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Enabled: false,
			APIKey:  "test-api-key-12345678",
		},
	}
	db, err := database.Open(config.DatabaseConfig{Driver: "sqlite", DSN: ":memory:"})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS config (key TEXT PRIMARY KEY, value TEXT NOT NULL);
		INSERT OR IGNORE INTO config (key, value) VALUES ('api_key', '');
	`)
	if err != nil {
		t.Fatalf("create config table: %v", err)
	}
	r := chi.NewRouter()
	auth.RegisterRoutes(r, cfg, db)
	return r, cfg
}

func TestLoginAuthDisabled(t *testing.T) {
	r, _ := setupTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["token"] == "" {
		t.Error("expected non-empty token")
	}
}

func TestMeValidToken(t *testing.T) {
	r, _ := setupTestRouter(t)

	loginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	var loginResp map[string]string
	if err := json.NewDecoder(loginW.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}
	token := loginResp["token"]
	if token == "" {
		t.Fatal("expected non-empty token from login")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meW := httptest.NewRecorder()
	r.ServeHTTP(meW, meReq)

	if meW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", meW.Code, meW.Body.String())
	}
	var meResp map[string]any
	if err := json.NewDecoder(meW.Body).Decode(&meResp); err != nil {
		t.Fatal(err)
	}
	if meResp["username"] != "admin" {
		t.Errorf("expected username 'admin', got %v", meResp["username"])
	}
}

func TestMeInvalidToken(t *testing.T) {
	r, _ := setupTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
