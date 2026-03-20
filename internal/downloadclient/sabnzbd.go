package downloadclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type sabnzbdConfig struct {
	Host     string
	Port     int
	UseSsl   bool
	APIKey   string
	Category string
}

type sabnzbdClient struct {
	cfg     sabnzbdConfig
	baseURL string
	http    *http.Client
}

func newSabnzbdClient(cfg sabnzbdConfig) *sabnzbdClient {
	scheme := "http"
	if cfg.UseSsl {
		scheme = "https"
	}
	return &sabnzbdClient{
		cfg:     cfg,
		baseURL: fmt.Sprintf("%s://%s:%d", scheme, cfg.Host, cfg.Port),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *sabnzbdClient) Name() string     { return "SABnzbd" }
func (c *sabnzbdClient) Protocol() string { return "usenet" }

func (c *sabnzbdClient) apiURL(mode string, extra url.Values) string {
	params := url.Values{
		"mode":   {mode},
		"output": {"json"},
		"apikey": {c.cfg.APIKey},
	}
	for k, v := range extra {
		params[k] = v
	}
	return fmt.Sprintf("%s/api?%s", c.baseURL, params.Encode())
}

func (c *sabnzbdClient) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL("version", nil), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sabnzbd version check failed: HTTP %d", resp.StatusCode)
	}
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("sabnzbd response parse: %w", err)
	}
	return nil
}

func (c *sabnzbdClient) AddTorrent(_ context.Context, _ string, _ string) error {
	return fmt.Errorf("SABnzbd does not support torrents")
}

func (c *sabnzbdClient) AddNZB(ctx context.Context, nzbURL string, category string) error {
	extra := url.Values{"name": {nzbURL}}
	cat := category
	if cat == "" {
		cat = c.cfg.Category
	}
	if cat != "" {
		extra.Set("cat", cat)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL("addurl", extra), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sabnzbd addurl failed: HTTP %d", resp.StatusCode)
	}
	var result struct {
		Status bool   `json:"status"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("sabnzbd response parse: %w", err)
	}
	if !result.Status {
		return fmt.Errorf("sabnzbd addurl error: %s", result.Error)
	}
	return nil
}

type sabSlot struct {
	NZOID    string `json:"nzo_id"`
	Filename string `json:"filename"`
	MB       string `json:"mb"`
	MBLeft   string `json:"mbleft"`
	Status   string `json:"status"`
	Cat      string `json:"cat"`
}

func (c *sabnzbdClient) GetItems(ctx context.Context) ([]Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL("queue", nil), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sabnzbd queue failed: HTTP %d", resp.StatusCode)
	}
	var result struct {
		Queue struct {
			Slots []sabSlot `json:"slots"`
		} `json:"queue"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("sabnzbd queue decode: %w", err)
	}
	items := make([]Item, len(result.Queue.Slots))
	for i, s := range result.Queue.Slots {
		mbTotal, _ := strconv.ParseFloat(s.MB, 64)
		mbLeft, _ := strconv.ParseFloat(s.MBLeft, 64)
		items[i] = Item{
			ID:       s.NZOID,
			Name:     s.Filename,
			Size:     int64(mbTotal * 1024 * 1024),
			SizeLeft: int64(mbLeft * 1024 * 1024),
			Status:   s.Status,
			Category: s.Cat,
		}
	}
	return items, nil
}

func (c *sabnzbdClient) RemoveItem(ctx context.Context, id string, deleteFiles bool) error {
	extra := url.Values{"id": {id}}
	if deleteFiles {
		extra.Set("del_files", "1")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL("delete", extra), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sabnzbd delete failed: HTTP %d", resp.StatusCode)
	}
	return nil
}
