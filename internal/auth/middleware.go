package auth

import (
"crypto/subtle"
"net/http"
)

func APIKeyMiddleware(apiKey string, enabled bool) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if !enabled || apiKey == "" {
next.ServeHTTP(w, r)
return
}
key := r.Header.Get("X-Api-Key")
if key == "" {
key = r.URL.Query().Get("apikey")
}
if subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) != 1 {
w.Header().Set("Content-Type", "application/json")
http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
return
}
next.ServeHTTP(w, r)
})
}
}

func BasicAuthMiddleware(username, password string, enabled bool) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if !enabled {
next.ServeHTTP(w, r)
return
}
user, pass, ok := r.BasicAuth()
if !ok ||
subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
w.Header().Set("WWW-Authenticate", `Basic realm="GoRadarr"`)
http.Error(w, `{"message":"Unauthorized"}`, http.StatusUnauthorized)
return
}
next.ServeHTTP(w, r)
})
}
}
