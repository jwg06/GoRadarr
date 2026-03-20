package downloadclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type nzbgetConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Category string
}

type nzbgetClient struct {
	cfg     nzbgetConfig
	baseURL string
	http    *http.Client
}

func newNzbgetClient(cfg nzbgetConfig) *nzbgetClient {
	return &nzbgetClient{
		cfg:     cfg,
		baseURL: fmt.Sprintf("http://%s:%d/jsonrpc", cfg.Host, cfg.Port),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *nzbgetClient) Name() string     { return "NZBGet" }
func (c *nzbgetClient) Protocol() string { return "usenet" }

type jsonRPCRequest struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
	ID     int    `json:"id"`
}

type jsonRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *nzbgetClient) call(ctx context.Context, method string, params []any, out any) error {
	body, err := json.Marshal(jsonRPCRequest{Method: method, Params: params, ID: 1})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.Username != "" {
		req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("nzbget rpc failed: HTTP %d", resp.StatusCode)
	}
	var rpcResp jsonRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("nzbget response decode: %w", err)
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("nzbget rpc error: %s", rpcResp.Error.Message)
	}
	if out != nil && rpcResp.Result != nil {
		return json.Unmarshal(rpcResp.Result, out)
	}
	return nil
}

func (c *nzbgetClient) TestConnection(ctx context.Context) error {
	var version string
	return c.call(ctx, "version", []any{}, &version)
}

func (c *nzbgetClient) AddTorrent(_ context.Context, _ string, _ string) error {
	return fmt.Errorf("NZBGet does not support torrents")
}

func (c *nzbgetClient) AddNZB(ctx context.Context, nzbURL string, category string) error {
	if category == "" {
		category = c.cfg.Category
	}
	// appendurl params: NZBFilename, Category, Priority, AddToTop, URL
	var nzbID int
	if err := c.call(ctx, "appendurl", []any{nzbURL, category, 0, false, nzbURL}, &nzbID); err != nil {
		return fmt.Errorf("nzbget appendurl: %w", err)
	}
	if nzbID == 0 {
		return fmt.Errorf("nzbget appendurl: server returned ID 0 (failure)")
	}
	return nil
}

type nzbgetGroup struct {
	NZBID           int    `json:"NZBID"`
	NZBName         string `json:"NZBName"`
	FileSizeMB      int64  `json:"FileSizeMB"`
	RemainingSizeMB int64  `json:"RemainingSizeMB"`
	Status          string `json:"Status"`
	Category        string `json:"Category"`
}

func (c *nzbgetClient) GetItems(ctx context.Context) ([]Item, error) {
	var groups []nzbgetGroup
	if err := c.call(ctx, "listgroups", []any{0}, &groups); err != nil {
		return nil, fmt.Errorf("nzbget listgroups: %w", err)
	}
	items := make([]Item, len(groups))
	for i, g := range groups {
		items[i] = Item{
			ID:       fmt.Sprintf("%d", g.NZBID),
			Name:     g.NZBName,
			Size:     g.FileSizeMB * 1024 * 1024,
			SizeLeft: g.RemainingSizeMB * 1024 * 1024,
			Status:   g.Status,
			Category: g.Category,
		}
	}
	return items, nil
}

func (c *nzbgetClient) RemoveItem(ctx context.Context, id string, _ bool) error {
	var nzbID int
	fmt.Sscanf(id, "%d", &nzbID)
	var success bool
	if err := c.call(ctx, "editqueue", []any{"GroupDelete", "", []int{nzbID}}, &success); err != nil {
		return fmt.Errorf("nzbget editqueue: %w", err)
	}
	return nil
}
