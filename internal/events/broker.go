package events

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type EventType string

const (
	EventMovieAdded       EventType = "movie.added"
	EventMovieUpdated     EventType = "movie.updated"
	EventMovieDeleted     EventType = "movie.deleted"
	EventQueueChanged     EventType = "queue.changed"
	EventTaskStarted      EventType = "task.started"
	EventTaskCompleted    EventType = "task.completed"
	EventHealthChanged    EventType = "health.changed"
	EventDownloadGrabbed  EventType = "download.grabbed"
	EventDownloadImported EventType = "download.imported"
)

type Event struct {
	Type EventType `json:"eventType"`
	Data any       `json:"data"`
}

type Broker struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
	logger  *slog.Logger
}

var (
	defaultBrokerMu sync.RWMutex
	defaultBroker   *Broker
)

func NewBroker(logger *slog.Logger) *Broker {
	return &Broker{clients: make(map[chan Event]struct{}), logger: logger}
}

func SetDefaultBroker(b *Broker) {
	defaultBrokerMu.Lock()
	defer defaultBrokerMu.Unlock()
	defaultBroker = b
}

func PublishDefault(evt Event) {
	defaultBrokerMu.RLock()
	broker := defaultBroker
	defaultBrokerMu.RUnlock()
	if broker != nil {
		broker.Publish(evt)
	}
}

func (b *Broker) Publish(evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- evt:
		default:
		}
	}
}

// Subscribe registers a channel that receives all published events.
// The caller must call Unsubscribe when done to avoid a resource leak.
func (b *Broker) Subscribe() chan Event {
	ch := make(chan Event, 64)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel and drains it.
func (b *Broker) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
	for len(ch) > 0 {
		<-ch
	}
}

func (b *Broker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := make(chan Event, 32)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	defer func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
		close(ch)
	}()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(evt)
			if err != nil {
				if b.logger != nil {
					b.logger.Warn("failed to marshal event", "error", err)
				}
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
