package server

import (
	"fmt"
	"net/http"
)

func openAPISpecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := embeddedFS.ReadFile("docs/openapi.yaml")
		if err != nil {
			http.Error(w, fmt.Sprintf("openapi spec unavailable: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
