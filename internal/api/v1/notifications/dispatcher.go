package notifications

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jwg06/goradarr/internal/database"
	"github.com/jwg06/goradarr/internal/events"
)

// Dispatcher subscribes to the SSE event broker and fires configured
// notification providers for matching event types.
type Dispatcher struct {
	db     *database.DB
	broker *events.Broker
	logger *slog.Logger
}

// NewDispatcher creates a Dispatcher. Call Start(ctx) to begin listening.
func NewDispatcher(db *database.DB, broker *events.Broker, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{db: db, broker: broker, logger: logger}
}

// Start subscribes to the broker and dispatches notifications until ctx is
// cancelled.  It is safe to call in a goroutine.
func (d *Dispatcher) Start(ctx context.Context) {
	ch := d.broker.Subscribe()
	defer d.broker.Unsubscribe(ch)

	d.logger.Info("notification dispatcher started")
	for {
		select {
		case <-ctx.Done():
			d.logger.Info("notification dispatcher stopped")
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			d.dispatch(ctx, evt)
		}
	}
}

// dispatch fires all matching notifiers for a single event.
func (d *Dispatcher) dispatch(ctx context.Context, evt events.Event) {
	col, ok := eventColumn(evt.Type)
	if !ok {
		return // event type has no notification trigger
	}

	query := `SELECT id, name, implementation, config_contract,
	                 COALESCE(settings,'{}'),
	                 on_grab, on_download, on_upgrade, on_rename, on_health_issue, on_delete,
	                 COALESCE(tags,'[]')
	          FROM notification_configs WHERE ` + col + ` = TRUE`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		d.logger.Error("dispatcher: query notification_configs", "err", err)
		return
	}
	defer rows.Close()

	payload := eventPayload(evt)

	for rows.Next() {
		var nc NotificationConfig
		if err := rows.Scan(
			&nc.ID, &nc.Name, &nc.Implementation, &nc.ConfigContract, &nc.Fields,
			&nc.OnGrab, &nc.OnDownload, &nc.OnUpgrade, &nc.OnRename, &nc.OnHealthIssue, &nc.OnDelete,
			&nc.Tags,
		); err != nil {
			d.logger.Warn("dispatcher: scan row", "err", err)
			continue
		}

		notifier, err := BuildNotifier(nc)
		if err != nil {
			d.logger.Warn("dispatcher: build notifier", "name", nc.Name, "err", err)
			continue
		}

		go func(name string, n Notifier) {
			if err := n.Send(ctx, string(evt.Type), payload); err != nil {
				d.logger.Warn("notification send failed", "name", name, "event", evt.Type, "err", err)
			} else {
				d.logger.Info("notification sent", "name", name, "event", evt.Type)
			}
		}(nc.Name, notifier)
	}
}

// eventColumn maps an EventType to the boolean column that gates it.
func eventColumn(t events.EventType) (string, bool) {
	switch t {
	case events.EventDownloadGrabbed:
		return "on_grab", true
	case events.EventDownloadImported:
		return "on_download", true
	case events.EventMovieDeleted:
		return "on_delete", true
	case events.EventHealthChanged:
		return "on_health_issue", true
	case events.EventMovieUpdated:
		return "on_rename", true
	default:
		return "", false
	}
}

// eventPayload converts an Event's Data to a flat map for notifiers.
func eventPayload(evt events.Event) map[string]any {
	if evt.Data == nil {
		return map[string]any{"eventType": string(evt.Type)}
	}
	// Marshal → unmarshal to get a plain map regardless of concrete type.
	b, err := json.Marshal(evt.Data)
	if err != nil {
		return map[string]any{"eventType": string(evt.Type)}
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]any{"eventType": string(evt.Type)}
	}
	m["eventType"] = string(evt.Type)
	return m
}

// BuildNotifier creates a Notifier from a persisted NotificationConfig.
// The settings JSON is parsed to populate provider-specific fields.
func BuildNotifier(nc NotificationConfig) (Notifier, error) {
	var fields map[string]string
	if err := json.Unmarshal(nc.Fields, &fields); err != nil {
		fields = map[string]string{}
	}

	switch nc.Implementation {
	case "Discord":
		return &DiscordNotifier{WebhookURL: fields["webhookUrl"], Username: fields["username"]}, nil
	case "Slack":
		return &SlackNotifier{WebhookURL: fields["webhookUrl"], Channel: fields["channel"], Username: fields["username"]}, nil
	case "Webhook":
		return &WebhookNotifier{URL: fields["url"], Method: fields["method"]}, nil
	case "Email":
		return &EmailNotifier{
			Server:   fields["server"],
			Port:     25,
			From:     fields["from"],
			To:       []string{fields["to"]},
			Username: fields["username"],
			Password: fields["password"],
		}, nil
	default:
		return &WebhookNotifier{URL: fields["url"], Method: "POST"}, nil
	}
}
