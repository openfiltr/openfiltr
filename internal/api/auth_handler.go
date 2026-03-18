package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	var id, hash, role string
	if err := h.db.QueryRow("SELECT id,password_hash,role FROM users WHERE username=?", req.Username).Scan(&id, &hash, &role); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !auth.CheckPassword(req.Password, hash) {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := h.authSvc.GenerateToken(id, req.Username, role)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "openfiltr_token", Value: token, Path: "/",
		HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
		MaxAge: h.cfg.Auth.TokenExpiry * 3600,
	})
	respond(w, http.StatusOK, map[string]string{"token": token, "username": req.Username, "role": role})
}

func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	var count int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
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
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	_, err = h.db.Exec(`INSERT INTO users(id,username,email,password_hash,role) VALUES(?,?,?,?,'admin')`,
		uuid.New().String(), req.Username, req.Username+"@localhost", hash)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	respond(w, http.StatusCreated, map[string]string{"message": "setup complete"})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "openfiltr_token", Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	})
	respond(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	respond(w, http.StatusOK, map[string]string{"id": c.UserID, "username": c.Username, "role": c.Role})
}

func (h *Handler) ListTokens(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	rows, err := h.db.Query(`SELECT id,name,scopes,last_used_at,expires_at,created_at FROM api_tokens WHERE user_id=? ORDER BY created_at DESC`, c.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	type item struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Scopes     string  `json:"scopes"`
		LastUsedAt *string `json:"last_used_at"`
		ExpiresAt  *string `json:"expires_at"`
		CreatedAt  string  `json:"created_at"`
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.ID, &it.Name, &it.Scopes, &it.LastUsedAt, &it.ExpiresAt, &it.CreatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	if items == nil {
		items = []item{}
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
	id := uuid.New().String()
	var exp interface{}
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid expires_at (use RFC3339)")
			return
		}
		exp = t.UTC().Format("2006-01-02 15:04:05")
	}
	if _, err := h.db.Exec(`INSERT INTO api_tokens(id,user_id,name,token_hash,scopes,expires_at) VALUES(?,?,?,?,'[]',?)`,
		id, c.UserID, req.Name, hash, exp); err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	respond(w, http.StatusCreated, map[string]string{"id": id, "name": req.Name, "token": raw})
}

func (h *Handler) DeleteToken(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	id := chi.URLParam(r, "id")
	res, err := h.db.Exec("DELETE FROM api_tokens WHERE id=? AND user_id=?", id, c.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "token not found")
		return
	}
	respond(w, http.StatusOK, map[string]string{"message": "deleted"})
}
