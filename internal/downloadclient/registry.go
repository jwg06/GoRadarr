package downloadclient

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Build creates a Client from a stored download_clients row.
// clientType is the implementation field (e.g. "qBittorrent", "SABnzbd", "NzbGet").
// fields is the JSON settings column from the download_clients table.
func Build(clientType string, fields json.RawMessage) (Client, error) {
	f := parseFields(fields)
	switch strings.ToLower(clientType) {
	case "qbittorrent":
		cfg := qbittorrentConfig{
			Host:     fieldString(f, "host"),
			Port:     fieldInt(f, "port"),
			UseSsl:   fieldBool(f, "useSsl"),
			Username: fieldString(f, "username"),
			Password: fieldString(f, "password"),
			Category: fieldString(f, "movieCategory"),
		}
		if cfg.Host == "" {
			cfg.Host = "localhost"
		}
		if cfg.Port == 0 {
			cfg.Port = 8080
		}
		return newQbittorrentClient(cfg), nil

	case "sabnzbd":
		cfg := sabnzbdConfig{
			Host:     fieldString(f, "host"),
			Port:     fieldInt(f, "port"),
			UseSsl:   fieldBool(f, "useSsl"),
			APIKey:   fieldString(f, "apiKey"),
			Category: fieldString(f, "movieCategory"),
		}
		if cfg.Host == "" {
			cfg.Host = "localhost"
		}
		if cfg.Port == 0 {
			cfg.Port = 8080
		}
		return newSabnzbdClient(cfg), nil

	case "nzbget":
		cfg := nzbgetConfig{
			Host:     fieldString(f, "host"),
			Port:     fieldInt(f, "port"),
			Username: fieldString(f, "username"),
			Password: fieldString(f, "password"),
			Category: fieldString(f, "movieCategory"),
		}
		if cfg.Host == "" {
			cfg.Host = "localhost"
		}
		if cfg.Port == 0 {
			cfg.Port = 6789
		}
		return newNzbgetClient(cfg), nil

	default:
		return nil, fmt.Errorf("unsupported download client type: %s", clientType)
	}
}

// parseFields parses Radarr-style settings stored as either a JSON array
// ([{"name":"key","value":"val"}]) or a plain JSON object ({"key":"val"}).
func parseFields(raw json.RawMessage) map[string]json.RawMessage {
	var arr []struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
		result := make(map[string]json.RawMessage, len(arr))
		for _, f := range arr {
			result[f.Name] = f.Value
		}
		return result
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj
	}
	return map[string]json.RawMessage{}
}

func fieldString(fields map[string]json.RawMessage, key string) string {
	v, ok := fields[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		return s
	}
	return strings.Trim(string(v), `"`)
}

func fieldInt(fields map[string]json.RawMessage, key string) int {
	v, ok := fields[key]
	if !ok {
		return 0
	}
	var n int
	if err := json.Unmarshal(v, &n); err == nil {
		return n
	}
	var f float64
	if err := json.Unmarshal(v, &f); err == nil {
		return int(f)
	}
	return 0
}

func fieldBool(fields map[string]json.RawMessage, key string) bool {
	v, ok := fields[key]
	if !ok {
		return false
	}
	var b bool
	if err := json.Unmarshal(v, &b); err == nil {
		return b
	}
	return false
}
