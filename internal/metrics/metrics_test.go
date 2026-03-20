package metrics_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwg06/goradarr/internal/metrics"
)

func TestRequestMiddlewareIncrementsCounter(t *testing.T) {
	before := metrics.RequestsTotal.Load()

	handler := metrics.RequestMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	after := metrics.RequestsTotal.Load()
	if after != before+1 {
		t.Errorf("expected RequestsTotal=%d, got %d", before+1, after)
	}
}

func TestHandlerReturnsExpectedKeys(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	for _, key := range []string{
		"goradarr_requests_total",
		"goradarr_request_errors",
		"goradarr_db_queries_total",
		"goradarr_queue_depth",
		"goradarr_uptime_seconds",
	} {
		if _, ok := body[key]; !ok {
			t.Errorf("expected key %q in metrics response", key)
		}
	}
}
