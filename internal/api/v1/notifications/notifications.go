package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/database"
)

// NotificationConfig holds persisted notification settings.
type NotificationConfig struct {
	ID             int64           `json:"id"`
	Name           string          `json:"name"`
	Implementation string          `json:"implementation"`
	ConfigContract string          `json:"configContract"`
	Fields         json.RawMessage `json:"fields"`
	OnGrab         bool            `json:"onGrab"`
	OnDownload     bool            `json:"onDownload"`
	OnUpgrade      bool            `json:"onUpgrade"`
	OnRename       bool            `json:"onRename"`
	OnHealthIssue  bool            `json:"onHealthIssue"`
	OnDelete       bool            `json:"onDelete"`
	Tags           json.RawMessage `json:"tags"`
}

// Notifier is the interface each notification provider must implement.
type Notifier interface {
	Send(ctx context.Context, event string, payload map[string]any) error
	Test(ctx context.Context) error
}

type handler struct{ db *database.DB }

func RegisterRoutes(r chi.Router, db *database.DB) {
	h := &handler{db: db}
	r.Route("/notification", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
		r.Post("/test", h.testAll)
		r.Post("/{id}/test", h.testOne)
	})
	r.Get("/notification/schema", h.schema)
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), `
		SELECT id, name, implementation, config_contract, COALESCE(settings,'{}'),
		       on_grab, on_download, on_upgrade, on_rename, on_health_issue, on_delete,
		       COALESCE(tags,'[]')
		FROM notification_configs ORDER BY name`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var result []NotificationConfig
	for rows.Next() {
		var n NotificationConfig
		rows.Scan(&n.ID, &n.Name, &n.Implementation, &n.ConfigContract, &n.Fields,
			&n.OnGrab, &n.OnDownload, &n.OnUpgrade, &n.OnRename, &n.OnHealthIssue, &n.OnDelete, &n.Tags)
		result = append(result, n)
	}
	if result == nil {
		result = []NotificationConfig{}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var n NotificationConfig
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, name, implementation, config_contract, COALESCE(settings,'{}'),
		       on_grab, on_download, on_upgrade, on_rename, on_health_issue, on_delete,
		       COALESCE(tags,'[]')
		FROM notification_configs WHERE id=?`, id).Scan(
		&n.ID, &n.Name, &n.Implementation, &n.ConfigContract, &n.Fields,
		&n.OnGrab, &n.OnDownload, &n.OnUpgrade, &n.OnRename, &n.OnHealthIssue, &n.OnDelete, &n.Tags)
	if err != nil {
		writeError(w, http.StatusNotFound, "notification not found")
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var n NotificationConfig
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if n.Fields == nil {
		n.Fields = json.RawMessage("{}")
	}
	if n.Tags == nil {
		n.Tags = json.RawMessage("[]")
	}
	res, err := h.db.ExecContext(r.Context(), `
		INSERT INTO notification_configs (name, implementation, config_contract, settings,
		    on_grab, on_download, on_upgrade, on_rename, on_health_issue, on_delete, tags)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		n.Name, n.Implementation, n.ConfigContract, string(n.Fields),
		n.OnGrab, n.OnDownload, n.OnUpgrade, n.OnRename, n.OnHealthIssue, n.OnDelete, string(n.Tags))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	n.ID, _ = res.LastInsertId()
	writeJSON(w, http.StatusCreated, n)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var n NotificationConfig
	json.NewDecoder(r.Body).Decode(&n)
	n.ID = id
	h.db.ExecContext(r.Context(), `
		UPDATE notification_configs SET name=?, implementation=?, config_contract=?, settings=?,
		    on_grab=?, on_download=?, on_upgrade=?, on_rename=?, on_health_issue=?, on_delete=?,
		    tags=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		n.Name, n.Implementation, n.ConfigContract, string(n.Fields),
		n.OnGrab, n.OnDownload, n.OnUpgrade, n.OnRename, n.OnHealthIssue, n.OnDelete,
		string(n.Tags), n.ID)
	writeJSON(w, http.StatusOK, n)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.db.ExecContext(r.Context(), "DELETE FROM notification_configs WHERE id=?", id)
	w.WriteHeader(http.StatusOK)
}

func (h *handler) testAll(w http.ResponseWriter, r *http.Request)  { writeJSON(w, 200, []any{}) }
func (h *handler) testOne(w http.ResponseWriter, r *http.Request) {
	// TODO: instantiate the notifier and call Test()
	w.WriteHeader(http.StatusOK)
}

func (h *handler) schema(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, NotificationSchemas())
}

// --- Provider implementations ---

// DiscordNotifier sends webhook notifications to Discord.
type DiscordNotifier struct {
	WebhookURL string
	Username   string
	AvatarURL  string
}

func (d *DiscordNotifier) Send(ctx context.Context, event string, payload map[string]any) error {
	body := map[string]any{
		"username":   d.Username,
		"avatar_url": d.AvatarURL,
		"embeds": []map[string]any{
			{
				"title":       fmt.Sprintf("GoRadarr — %s", event),
				"description": fmt.Sprintf("%v", payload["title"]),
				"color":       3447003,
			},
		},
	}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, d.WebhookURL, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook status %d", resp.StatusCode)
	}
	return nil
}

func (d *DiscordNotifier) Test(ctx context.Context) error {
	return d.Send(ctx, "Test", map[string]any{"title": "GoRadarr test notification"})
}

// SlackNotifier sends webhook notifications to Slack.
type SlackNotifier struct {
	WebhookURL string
	Channel    string
	Username   string
}

func (s *SlackNotifier) Send(ctx context.Context, event string, payload map[string]any) error {
	body := map[string]any{
		"channel":  s.Channel,
		"username": s.Username,
		"text":     fmt.Sprintf("*%s*: %v", event, payload["title"]),
	}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.WebhookURL, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *SlackNotifier) Test(ctx context.Context) error {
	return s.Send(ctx, "Test", map[string]any{"title": "GoRadarr test notification"})
}

// WebhookNotifier sends a generic JSON POST to a URL.
type WebhookNotifier struct {
	URL    string
	Method string
}

func (wh *WebhookNotifier) Send(ctx context.Context, event string, payload map[string]any) error {
	payload["eventType"] = event
	data, _ := json.Marshal(payload)
	method := wh.Method
	if method == "" {
		method = http.MethodPost
	}
	req, _ := http.NewRequestWithContext(ctx, method, wh.URL, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (wh *WebhookNotifier) Test(ctx context.Context) error {
	return wh.Send(ctx, "Test", map[string]any{"title": "GoRadarr test notification"})
}

// EmailNotifier sends email via SMTP.
type EmailNotifier struct {
	Server   string
	Port     int
	From     string
	To       []string
	Username string
	Password string
}

func (e *EmailNotifier) Send(ctx context.Context, event string, payload map[string]any) error {
	addr := fmt.Sprintf("%s:%d", e.Server, e.Port)
	auth := smtp.PlainAuth("", e.Username, e.Password, e.Server)
	subject := fmt.Sprintf("GoRadarr — %s: %v", event, payload["title"])
	body := fmt.Sprintf("Subject: %s\r\nFrom: %s\r\nTo: %s\r\n\r\n%s",
		subject, e.From, strings.Join(e.To, ","), subject)
	return smtp.SendMail(addr, auth, e.From, e.To, []byte(body))
}

func (e *EmailNotifier) Test(ctx context.Context) error {
	return e.Send(ctx, "Test", map[string]any{"title": "GoRadarr test notification"})
}

func NotificationSchemas() []map[string]any {
	return []map[string]any{
		{
			"implementation": "Discord",
			"configContract": "DiscordSettings",
			"fields": []map[string]any{
				{"name": "webHookUrl", "label": "Webhook URL", "type": "url"},
				{"name": "username", "label": "Username", "type": "textbox", "value": "GoRadarr"},
				{"name": "avatar", "label": "Avatar URL", "type": "textbox"},
			},
		},
		{
			"implementation": "Slack",
			"configContract": "SlackSettings",
			"fields": []map[string]any{
				{"name": "webHookUrl", "label": "Webhook URL", "type": "url"},
				{"name": "channel", "label": "Channel", "type": "textbox"},
				{"name": "username", "label": "Username", "type": "textbox", "value": "GoRadarr"},
			},
		},
		{
			"implementation": "Webhook",
			"configContract": "WebhookSettings",
			"fields": []map[string]any{
				{"name": "url", "label": "URL", "type": "url"},
				{"name": "method", "label": "Method", "type": "select", "value": "POST"},
			},
		},
		{
			"implementation": "Email",
			"configContract": "EmailSettings",
			"fields": []map[string]any{
				{"name": "server", "label": "Server", "type": "textbox"},
				{"name": "port", "label": "Port", "type": "number", "value": 587},
				{"name": "from", "label": "From Address", "type": "textbox"},
				{"name": "to", "label": "To Address(es)", "type": "textbox"},
				{"name": "username", "label": "Username", "type": "textbox"},
				{"name": "password", "label": "Password", "type": "password"},
			},
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"message": msg})
}
