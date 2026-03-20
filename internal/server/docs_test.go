package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAPISpecHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	w := httptest.NewRecorder()

	openAPISpecHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Body.String(); len(got) == 0 || !strings.HasPrefix(got, "openapi") {
		t.Fatalf("expected openapi yaml, got %q", got)
	}
}

func TestDocsFSRoot(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	docsFS().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if body == "" || !(strings.Contains(body, "SwaggerUIBundle") || strings.Contains(body, "GoRadarr API Docs")) {
		t.Fatalf("unexpected docs body: %q", body)
	}
}
