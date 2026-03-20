package metrics

import (
	"expvar"
	"net/http"
	"sync/atomic"
	"time"
)

// Counters exposed via expvar.
var (
	RequestsTotal  atomic.Int64
	RequestErrors  atomic.Int64
	DBQueriesTotal atomic.Int64
	QueueDepth     atomic.Int64
	StartTime      = time.Now()
)

func init() {
	expvar.Publish("goradarr_requests_total", expvar.Func(func() any { return RequestsTotal.Load() }))
	expvar.Publish("goradarr_request_errors", expvar.Func(func() any { return RequestErrors.Load() }))
	expvar.Publish("goradarr_db_queries_total", expvar.Func(func() any { return DBQueriesTotal.Load() }))
	expvar.Publish("goradarr_queue_depth", expvar.Func(func() any { return QueueDepth.Load() }))
	expvar.Publish("goradarr_uptime_seconds", expvar.Func(func() any { return int64(time.Since(StartTime).Seconds()) }))
}

// Handler returns metrics as JSON (expvar format).
func Handler() http.Handler { return expvar.Handler() }

// RequestMiddleware increments RequestsTotal per request.
func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RequestsTotal.Add(1)
		next.ServeHTTP(w, r)
	})
}
