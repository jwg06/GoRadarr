package logging

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// LogEntry is a single captured log record stored in the ring buffer.
type LogEntry struct {
	Time  string         `json:"time"`
	Level string         `json:"level"`
	Msg   string         `json:"msg"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

const ringCapacity = 500

type ringBuffer struct {
	mu      sync.RWMutex
	entries [ringCapacity]LogEntry
	head    int // next write position
	count   int // number of stored entries (0..ringCapacity)
	subsMu  sync.RWMutex
	subs    map[chan LogEntry]struct{}
}

var globalRing = &ringBuffer{
	subs: make(map[chan LogEntry]struct{}),
}

func (rb *ringBuffer) push(e LogEntry) {
	rb.mu.Lock()
	rb.entries[rb.head] = e
	rb.head = (rb.head + 1) % ringCapacity
	if rb.count < ringCapacity {
		rb.count++
	}
	rb.mu.Unlock()

	rb.subsMu.RLock()
	for ch := range rb.subs {
		select {
		case ch <- e:
		default: // subscriber too slow; drop
		}
	}
	rb.subsMu.RUnlock()
}

// recent returns the last n log entries in chronological order.
func (rb *ringBuffer) recent(n int) []LogEntry {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	if n > rb.count {
		n = rb.count
	}
	if n == 0 {
		return []LogEntry{}
	}
	// oldest slot: (head - count + large_multiple) % ringCapacity
	oldest := ((rb.head - rb.count) % ringCapacity + ringCapacity) % ringCapacity
	startOffset := rb.count - n
	result := make([]LogEntry, n)
	for i := 0; i < n; i++ {
		result[i] = rb.entries[(oldest+startOffset+i)%ringCapacity]
	}
	return result
}

func (rb *ringBuffer) subscribe() chan LogEntry {
	ch := make(chan LogEntry, 64)
	rb.subsMu.Lock()
	rb.subs[ch] = struct{}{}
	rb.subsMu.Unlock()
	return ch
}

func (rb *ringBuffer) unsubscribe(ch chan LogEntry) {
	rb.subsMu.Lock()
	delete(rb.subs, ch)
	rb.subsMu.Unlock()
	for len(ch) > 0 {
		<-ch
	}
}

// RecentLogs returns the last n log entries from the in-process ring buffer.
func RecentLogs(n int) []LogEntry {
	return globalRing.recent(n)
}

// Subscribe returns a channel that receives new log entries in real time.
func Subscribe() chan LogEntry {
	return globalRing.subscribe()
}

// Unsubscribe removes a subscriber channel and drains it.
func Unsubscribe(ch chan LogEntry) {
	globalRing.unsubscribe(ch)
}

// RingHandler returns a slog.Handler that captures records into the ring buffer.
// opts controls which levels are captured (nil → LevelInfo).
func RingHandler(opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ringHandler{opts: *opts}
}

type ringHandler struct {
	opts  slog.HandlerOptions
	attrs []slog.Attr
	group string
}

func (h *ringHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	if h.opts.Level == nil {
		return lvl >= slog.LevelInfo
	}
	return lvl >= h.opts.Level.Level()
}

func (h *ringHandler) Handle(_ context.Context, r slog.Record) error {
	t := r.Time
	if t.IsZero() {
		t = time.Now()
	}
	entry := LogEntry{
		Time:  t.Format(time.RFC3339),
		Level: r.Level.String(),
		Msg:   r.Message,
	}
	var attrs map[string]any
	for _, a := range h.attrs {
		if attrs == nil {
			attrs = make(map[string]any)
		}
		attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		if attrs == nil {
			attrs = make(map[string]any)
		}
		attrs[a.Key] = a.Value.Any()
		return true
	})
	entry.Attrs = attrs
	globalRing.push(entry)
	return nil
}

func (h *ringHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(merged, h.attrs)
	copy(merged[len(h.attrs):], attrs)
	return &ringHandler{opts: h.opts, attrs: merged, group: h.group}
}

func (h *ringHandler) WithGroup(name string) slog.Handler {
	return &ringHandler{opts: h.opts, attrs: h.attrs, group: name}
}

// teeHandler sends each log record to two handlers.
type teeHandler struct {
	primary slog.Handler
	ring    slog.Handler
}

func (t *teeHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return t.primary.Enabled(ctx, lvl) || t.ring.Enabled(ctx, lvl)
}

func (t *teeHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	if t.primary.Enabled(ctx, r.Level) {
		firstErr = t.primary.Handle(ctx, r)
	}
	if t.ring.Enabled(ctx, r.Level) {
		if err := t.ring.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (t *teeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &teeHandler{primary: t.primary.WithAttrs(attrs), ring: t.ring.WithAttrs(attrs)}
}

func (t *teeHandler) WithGroup(name string) slog.Handler {
	return &teeHandler{primary: t.primary.WithGroup(name), ring: t.ring.WithGroup(name)}
}
