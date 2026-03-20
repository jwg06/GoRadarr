package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jwg06/goradarr/internal/config"
	"github.com/jwg06/goradarr/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Expires string `json:"expires"`
}

type Handler struct {
	cfg *config.Config
	db  *database.DB
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	secret := JWTSecretFromAPIKey(h.cfg.Auth.APIKey)

	if !h.cfg.Auth.Enabled {
		token, err := GenerateToken(secret, "admin")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to generate token"})
			return
		}
		writeJSON(w, http.StatusOK, LoginResponse{
			Token:   token,
			Expires: time.Now().Add(tokenDuration).Format(time.RFC3339),
		})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid request"})
		return
	}

	if req.Username != h.cfg.Auth.Username {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(h.cfg.Auth.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "invalid credentials"})
		return
	}

	token, err := GenerateToken(secret, req.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to generate token"})
		return
	}
	writeJSON(w, http.StatusOK, LoginResponse{
		Token:   token,
		Expires: time.Now().Add(tokenDuration).Format(time.RFC3339),
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "missing token"})
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := ValidateToken(JWTSecretFromAPIKey(h.cfg.Auth.APIKey), tokenStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "invalid token"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"username":      claims.Username,
		"authenticated": true,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *Handler) RegenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	key, err := GenerateAPIKey()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to generate API key"})
		return
	}
	_, err = h.db.ExecContext(r.Context(),
		"INSERT OR REPLACE INTO config (key, value) VALUES ('api_key', ?)", key)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to store API key"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"apiKey": key})
}

func RegisterRoutes(r chi.Router, cfg *config.Config, db *database.DB) {
	h := &Handler{cfg: cfg, db: db}
	r.Post("/login", h.Login)
	r.Get("/me", h.Me)
	r.Post("/logout", h.Logout)
	r.Post("/apikey/regenerate", h.RegenerateAPIKey)
}
