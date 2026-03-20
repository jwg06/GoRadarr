package downloadclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type qbittorrentConfig struct {
	Host     string
	Port     int
	UseSsl   bool
	Username string
	Password string
	Category string
}

type qbittorrentClient struct {
	cfg     qbittorrentConfig
	baseURL string
	http    *http.Client
}

func newQbittorrentClient(cfg qbittorrentConfig) *qbittorrentClient {
	jar, _ := cookiejar.New(nil)
	scheme := "http"
	if cfg.UseSsl {
		scheme = "https"
	}
	return &qbittorrentClient{
		cfg:     cfg,
		baseURL: fmt.Sprintf("%s://%s:%d", scheme, cfg.Host, cfg.Port),
		http:    &http.Client{Timeout: 30 * time.Second, Jar: jar},
	}
}

func (c *qbittorrentClient) Name() string     { return "qBittorrent" }
func (c *qbittorrentClient) Protocol() string { return "torrent" }

func (c *qbittorrentClient) login(ctx context.Context) error {
	form := url.Values{
		"username": {c.cfg.Username},
		"password": {c.cfg.Password},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v2/auth/login", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if strings.TrimSpace(string(body)) != "Ok." {
		return fmt.Errorf("qbittorrent login failed: %s", string(body))
	}
	return nil
}

func (c *qbittorrentClient) TestConnection(ctx context.Context) error {
	if err := c.login(ctx); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/api/v2/app/version", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("qbittorrent version check failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *qbittorrentClient) AddTorrent(ctx context.Context, magnetOrURL string, savePath string) error {
	if err := c.login(ctx); err != nil {
		return err
	}
	form := url.Values{"urls": {magnetOrURL}}
	if savePath != "" {
		form.Set("savepath", savePath)
	}
	if c.cfg.Category != "" {
		form.Set("category", c.cfg.Category)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v2/torrents/add", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("qbittorrent add torrent failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *qbittorrentClient) AddNZB(_ context.Context, _ string, _ string) error {
	return fmt.Errorf("qBittorrent does not support NZB")
}

type qbTorrent struct {
	Hash     string `json:"hash"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	AmtLeft  int64  `json:"amount_left"`
	State    string `json:"state"`
	Category string `json:"category"`
}

func (c *qbittorrentClient) GetItems(ctx context.Context) ([]Item, error) {
	if err := c.login(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/api/v2/torrents/info", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("qbittorrent list torrents failed: HTTP %d", resp.StatusCode)
	}
	var torrents []qbTorrent
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, fmt.Errorf("qbittorrent decode torrents: %w", err)
	}
	items := make([]Item, len(torrents))
	for i, t := range torrents {
		items[i] = Item{
			ID:       t.Hash,
			Name:     t.Name,
			Size:     t.Size,
			SizeLeft: t.AmtLeft,
			Status:   t.State,
			Hash:     t.Hash,
			Category: t.Category,
		}
	}
	return items, nil
}

func (c *qbittorrentClient) RemoveItem(ctx context.Context, id string, deleteFiles bool) error {
	if err := c.login(ctx); err != nil {
		return err
	}
	del := "false"
	if deleteFiles {
		del = "true"
	}
	form := url.Values{
		"hashes":      {id},
		"deleteFiles": {del},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v2/torrents/delete", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("qbittorrent remove torrent failed: HTTP %d", resp.StatusCode)
	}
	return nil
}
