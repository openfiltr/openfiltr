package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/openfiltr/openfiltr/internal/auth"
	"github.com/openfiltr/openfiltr/internal/config"
	"github.com/openfiltr/openfiltr/internal/storage"
)

type Handler struct {
	db      storage.Store
	cfg     *config.Config
	authSvc *auth.Service
	version string
}

type contextKey string

const claimsKey contextKey = "claims"

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]interface{}{"error": map[string]string{"message": message}})
}

func decode(r *http.Request, v interface{}) error {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func queryInt(r *http.Request, key string, def int) int {
	if s := r.URL.Query().Get(key); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	return def
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
}

func (h *Handler) Version(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{"version": h.version})
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	var br, ar, cc int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM block_rules WHERE enabled=1").Scan(&br)
	_ = h.db.QueryRow("SELECT COUNT(*) FROM allow_rules WHERE enabled=1").Scan(&ar)
	_ = h.db.QueryRow("SELECT COUNT(*) FROM clients").Scan(&cc)
	respond(w, http.StatusOK, map[string]interface{}{
		"status": "running", "block_rule_count": br, "allow_rule_count": ar,
		"client_count": cc, "version": h.version,
	})
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		total, blocked, allowed, err := bolt.ActivityCounts()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		rate := 0.0
		if total > 0 {
			rate = float64(blocked) / float64(total) * 100
		}
		respond(w, http.StatusOK, map[string]interface{}{
			"total_queries": total, "blocked_queries": blocked, "allowed_queries": allowed,
			"block_rate": fmt.Sprintf("%.2f", rate),
		})
		return
	}

	var total, blocked, allowed int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM activity_log").Scan(&total)
	_ = h.db.QueryRow("SELECT COUNT(*) FROM activity_log WHERE action='blocked'").Scan(&blocked)
	_ = h.db.QueryRow("SELECT COUNT(*) FROM activity_log WHERE action='allowed'").Scan(&allowed)
	rate := 0.0
	if total > 0 {
		rate = float64(blocked) / float64(total) * 100
	}
	respond(w, http.StatusOK, map[string]interface{}{
		"total_queries": total, "blocked_queries": blocked, "allowed_queries": allowed,
		"block_rate": fmt.Sprintf("%.2f", rate),
	})
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := auth.ExtractToken(r)
		if token == "" {
			respondError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		claims, err := h.authSvc.ValidateToken(token)
		if err != nil {
			claims, err = h.authSvc.ValidateAPIToken(token)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
		}
		if bearerTokenFromRequest(r) == "" && isStateChanging(r.Method) {
			csrfCookie, err := r.Cookie(csrfCookieName)
			if err != nil || csrfCookie.Value == "" || r.Header.Get(csrfHeaderName) == "" || r.Header.Get(csrfHeaderName) != csrfCookie.Value {
				respondError(w, http.StatusForbidden, "CSRF token required")
				return
			}
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func currentUser(r *http.Request) *auth.Claims {
	c, _ := r.Context().Value(claimsKey).(*auth.Claims)
	return c
}
