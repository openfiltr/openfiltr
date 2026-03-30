package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/openfiltr/openfiltr/internal/auth"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.authSvc.LookupUserByUsername(req.Username)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := h.authSvc.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	csrfToken, err := newCSRFToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate CSRF token")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: authCookieName, Value: token, Path: "/",
		HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
		MaxAge: h.cfg.Auth.TokenExpiry * 3600,
	})
	setCSRFCookie(w, csrfToken, h.cfg.Auth.TokenExpiry*3600)
	w.Header().Set(csrfHeaderName, csrfToken)
	respond(w, http.StatusOK, map[string]string{"token": token, "username": user.Username, "role": user.Role})
}

func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	count, err := h.authSvc.CountUsers()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if count > 0 {
		respondError(w, http.StatusConflict, "setup already completed")
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "username required and password must be at least 8 characters")
		return
	}
	if err := h.authSvc.CreateAdminUser(req.Username, req.Password); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	respond(w, http.StatusCreated, map[string]string{"message": "setup complete"})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: authCookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	})
	setCSRFCookie(w, "", -1)
	respond(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	respond(w, http.StatusOK, map[string]string{"id": c.UserID, "username": c.Username, "role": c.Role})
}

func (h *Handler) ListTokens(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	items, err := h.authSvc.ListAPITokens(c.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": len(items)})
}

func (h *Handler) CreateToken(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	var req struct {
		Name      string  `json:"name"`
		ExpiresAt *string `json:"expires_at"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	raw, hash, err := auth.GenerateAPIToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	var exp *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid expires_at (use RFC3339)")
			return
		}
		exp = &t
	}
	id, err := h.authSvc.CreateAPIToken(c.UserID, req.Name, hash, exp)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	respond(w, http.StatusCreated, map[string]string{"id": id, "name": req.Name, "token": raw})
}

func (h *Handler) DeleteToken(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	id := chi.URLParam(r, "id")
	deleted, err := h.authSvc.DeleteAPIToken(id, c.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if !deleted {
		respondError(w, http.StatusNotFound, "token not found")
		return
	}
	respond(w, http.StatusOK, map[string]string{"message": "deleted"})
}
