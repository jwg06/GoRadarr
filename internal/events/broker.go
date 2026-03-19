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
EventMovieAdded      EventType = "movie.added"
EventMovieUpdated    EventType = "movie.updated"
EventMovieDeleted    EventType = "movie.deleted"
EventQueueChanged    EventType = "queue.changed"
EventTaskStarted     EventType = "task.started"
EventTaskCompleted   EventType = "task.completed"
EventHealthChanged   EventType = "health.changed"
EventDownloadGrabbed EventType = "download.grabbed"
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

func NewBroker(logger *slog.Logger) *Broker {
return &Broker{clients: make(map[chan Event]struct{}), logger: logger}
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
fmt.Fprintf(w, ": heartbeat\n\n")
flusher.Flush()
case evt, ok := <-ch:
if !ok {
return
}
data, _ := json.Marshal(evt)
fmt.Fprintf(w, "data: %s\n\n", data)
flusher.Flush()
}
}
}
