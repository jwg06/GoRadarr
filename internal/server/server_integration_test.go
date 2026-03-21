package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
)

// newTestServer creates a real httptest.Server backed by the full chi router
// and an in-memory SQLite database. Tests are hermetic and share nothing.
func newTestServer(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()

	db, err := database.Open(config.DatabaseConfig{Driver: "sqlite", DSN: ":memory:"})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	cfg := &config.Config{
		Host: "127.0.0.1",
		Port: 0,
		Auth: config.AuthConfig{
			Enabled: false,
			APIKey:  "test-api-key-abcdef1234567890abcd",
		},
		Data:      config.DataConfig{RootDir: t.TempDir()},
		Scheduler: config.SchedulerConfig{Enabled: false},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := New(cfg, db, logger)

	ts := httptest.NewServer(srv.http.Handler)
	t.Cleanup(func() {
		ts.Close()
		db.Close()
	})
	return ts, ts.Client()
}

// jsonBody marshals v and returns a reader suitable for an http request body.
func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return bytes.NewBuffer(b)
}

// decodeJSON decodes JSON from r into v, fatally failing on error.
func decodeJSON(t *testing.T, r io.Reader, v any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

// -------------------------------------------------------------------
// /api/v1/ping
// -------------------------------------------------------------------

func TestIntegration_Ping(t *testing.T) {
	ts, client := newTestServer(t)

	resp, err := client.Get(ts.URL + "/api/v1/ping")
	if err != nil {
		t.Fatalf("GET /api/v1/ping: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}

// -------------------------------------------------------------------
// /api/v1/system/status
// -------------------------------------------------------------------

func TestIntegration_SystemStatus(t *testing.T) {
	ts, client := newTestServer(t)

	resp, err := client.Get(ts.URL + "/api/v1/system/status")
	if err != nil {
		t.Fatalf("GET /api/v1/system/status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	decodeJSON(t, resp.Body, &body)
	if _, ok := body["version"]; !ok {
		t.Errorf("response missing 'version' field; got keys: %v", keys(body))
	}
}

// -------------------------------------------------------------------
// /api/v1/movie  (full CRUD)
// -------------------------------------------------------------------

func TestIntegration_MovieCRUD(t *testing.T) {
	ts, client := newTestServer(t)
	base := ts.URL + "/api/v1/movie"

	// POST – add
	addPayload := map[string]any{
		"title":  "Integration Test Movie",
		"tmdbId": 99901,
		"year":   2024,
	}
	resp, err := client.Post(base+"/", "application/json", jsonBody(t, addPayload))
	if err != nil {
		t.Fatalf("POST /api/v1/movie/: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/v1/movie/: want 201, got %d: %s", resp.StatusCode, body)
	}

	var added map[string]any
	if err := json.Unmarshal(body, &added); err != nil {
		t.Fatalf("unmarshal add response: %v", err)
	}
	id := int64(added["id"].(float64))
	if id == 0 {
		t.Fatal("expected non-zero movie ID in POST response")
	}

	// GET list
	resp, err = client.Get(base + "/")
	if err != nil {
		t.Fatalf("GET /api/v1/movie/: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/v1/movie/: want 200, got %d", resp.StatusCode)
	}
	var list []map[string]any
	decodeJSON(t, resp.Body, &list)
	if len(list) == 0 {
		t.Error("expected at least one movie in list after add")
	}

	// GET by id
	resp, err = client.Get(fmt.Sprintf("%s/%d", base, id))
	if err != nil {
		t.Fatalf("GET /api/v1/movie/%d: %v", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/v1/movie/%d: want 200, got %d", id, resp.StatusCode)
	}

	// PUT update (qualityProfileId must match a seeded profile; 1 = "Any")
	updatePayload := map[string]any{
		"title":            "Updated Integration Movie",
		"tmdbId":           99901,
		"year":             2024,
		"qualityProfileId": 1,
	}
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%d", base, id), jsonBody(t, updatePayload))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("PUT /api/v1/movie/%d: %v", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("PUT /api/v1/movie/%d: want 200, got %d", id, resp.StatusCode)
	}

	// DELETE
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", base, id), nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("DELETE /api/v1/movie/%d: %v", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("DELETE /api/v1/movie/%d: want 200, got %d", id, resp.StatusCode)
	}

	// Confirm deletion
	resp, err = client.Get(fmt.Sprintf("%s/%d", base, id))
	if err != nil {
		t.Fatalf("GET after DELETE: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET after DELETE: want 404, got %d", resp.StatusCode)
	}
}

// -------------------------------------------------------------------
// /api/v1/qualityprofile
// -------------------------------------------------------------------

func TestIntegration_QualityProfile(t *testing.T) {
	ts, client := newTestServer(t)

	resp, err := client.Get(ts.URL + "/api/v1/qualityprofile/")
	if err != nil {
		t.Fatalf("GET /api/v1/qualityprofile/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	var profiles []map[string]any
	decodeJSON(t, resp.Body, &profiles)
	if len(profiles) == 0 {
		t.Error("expected quality profiles seeded by migrations")
	}
}

// -------------------------------------------------------------------
// /api/v1/auth/login  (auth disabled → token without credentials)
// -------------------------------------------------------------------

func TestIntegration_AuthLogin(t *testing.T) {
	ts, client := newTestServer(t)

	resp, err := client.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("POST /api/v1/auth/login: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/v1/auth/login: want 200, got %d: %s", resp.StatusCode, body)
	}

	var loginResp struct {
		Token   string `json:"token"`
		Expires string `json:"expires"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if loginResp.Token == "" {
		t.Error("expected non-empty JWT token in login response")
	}
	if loginResp.Expires == "" {
		t.Error("expected non-empty expires in login response")
	}
}

// -------------------------------------------------------------------
// /api/v1/auth/me  (Bearer token)
// -------------------------------------------------------------------

func TestIntegration_AuthMe(t *testing.T) {
	ts, client := newTestServer(t)

	// Obtain a token first
	resp, err := client.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	decodeJSON(t, resp.Body, &loginResp)
	resp.Body.Close()

	if loginResp.Token == "" {
		t.Fatal("no token from login; cannot test /me")
	}

	// GET /me with the token
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/v1/auth/me: want 200, got %d: %s", resp.StatusCode, body)
	}

	var me map[string]any
	if err := json.Unmarshal(body, &me); err != nil {
		t.Fatalf("unmarshal /me response: %v", err)
	}
	if me["username"] == nil {
		t.Errorf("expected 'username' in /me response; got %v", me)
	}
}

// -------------------------------------------------------------------
// /api/v1/history
// -------------------------------------------------------------------

func TestIntegration_History(t *testing.T) {
	ts, client := newTestServer(t)

	resp, err := client.Get(ts.URL + "/api/v1/history/")
	if err != nil {
		t.Fatalf("GET /api/v1/history/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	// Response is a paged envelope with a "records" array.
	var envelope map[string]any
	decodeJSON(t, resp.Body, &envelope)
	if _, ok := envelope["records"]; !ok {
		t.Errorf("history response missing 'records' key; got %v", keys(envelope))
	}
}

// -------------------------------------------------------------------
// /api/v1/calendar
// -------------------------------------------------------------------

func TestIntegration_Calendar(t *testing.T) {
	ts, client := newTestServer(t)

	resp, err := client.Get(ts.URL + "/api/v1/calendar")
	if err != nil {
		t.Fatalf("GET /api/v1/calendar: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	// Must be a valid JSON array (possibly empty).
	var result []map[string]any
	decodeJSON(t, resp.Body, &result)
	// An empty slice is valid; just ensure decode succeeded.
}

// -------------------------------------------------------------------
// Auth: routes are open when auth.enabled = false
// -------------------------------------------------------------------

func TestIntegration_AuthDisabled_NoKeyRequired(t *testing.T) {
	ts, client := newTestServer(t)

	// Without any API key header, protected routes should still respond.
	tests := []string{
		"/api/v1/system/status",
		"/api/v1/movie/",
		"/api/v1/qualityprofile/",
		"/api/v1/history/",
		"/api/v1/calendar",
	}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			resp, err := client.Get(ts.URL + path)
			if err != nil {
				t.Fatalf("GET %s: %v", path, err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("GET %s: want 200, got %d", path, resp.StatusCode)
			}
		})
	}
}

// keys returns the map keys as a slice for diagnostic messages.
func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
