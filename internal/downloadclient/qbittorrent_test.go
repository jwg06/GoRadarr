package downloadclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func newTestQbClient(t *testing.T, srv *httptest.Server) *qbittorrentClient {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(u.Port())
	return newQbittorrentClient(qbittorrentConfig{
		Host:     u.Hostname(),
		Port:     port,
		Username: "admin",
		Password: "password",
	})
}

func TestQbittorrentTestConnection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			w.Write([]byte("Ok."))
		case "/api/v2/app/version":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"version": "5.0.0"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestQbClient(t, srv)
	if err := c.TestConnection(context.Background()); err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}
}

func TestQbittorrentGetItems(t *testing.T) {
	torrents := []qbTorrent{
		{
			Hash:     "abc123",
			Name:     "Test.Movie.2024.1080p",
			Size:     2_000_000_000,
			AmtLeft:  1_000_000_000,
			State:    "downloading",
			Category: "radarr",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			w.Write([]byte("Ok."))
		case "/api/v2/torrents/info":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(torrents)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestQbClient(t, srv)
	items, err := c.GetItems(context.Background())
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "abc123" {
		t.Errorf("expected ID abc123, got %s", items[0].ID)
	}
	if items[0].Name != "Test.Movie.2024.1080p" {
		t.Errorf("unexpected name: %s", items[0].Name)
	}
	if items[0].Size != 2_000_000_000 {
		t.Errorf("unexpected size: %d", items[0].Size)
	}
	if items[0].SizeLeft != 1_000_000_000 {
		t.Errorf("unexpected size_left: %d", items[0].SizeLeft)
	}
	if items[0].Status != "downloading" {
		t.Errorf("unexpected status: %s", items[0].Status)
	}
}

func TestQbittorrentLoginFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			w.Write([]byte("Fails."))
		}
	}))
	defer srv.Close()

	c := newTestQbClient(t, srv)
	if err := c.TestConnection(context.Background()); err == nil {
		t.Fatal("expected login error, got nil")
	}
}

func TestBuildQbittorrent(t *testing.T) {
	fields := []byte(`[{"name":"host","value":"localhost"},{"name":"port","value":8080},{"name":"useSsl","value":false}]`)
	client, err := Build("qBittorrent", fields)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if client.Name() != "qBittorrent" {
		t.Errorf("unexpected name: %s", client.Name())
	}
	if client.Protocol() != "torrent" {
		t.Errorf("unexpected protocol: %s", client.Protocol())
	}
}

func TestBuildUnsupported(t *testing.T) {
	_, err := Build("Deluge", []byte(`{}`))
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
