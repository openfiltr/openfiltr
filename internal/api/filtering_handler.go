package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ---- helpers shared by block_rules and allow_rules ----

type rule struct {
	ID        string  `json:"id"`
	Pattern   string  `json:"pattern"`
	RuleType  string  `json:"rule_type"`
	Comment   *string `json:"comment"`
	Enabled   int     `json:"enabled"`
	CreatedBy *string `json:"created_by"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func (h *Handler) listRules(w http.ResponseWriter, r *http.Request, table string) {
	limit := queryInt(r, "limit", 100)
	offset := queryInt(r, "offset", 0)
	var total int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM "+table).Scan(&total)
	rows, err := h.db.Query("SELECT id,pattern,rule_type,comment,enabled,created_by,created_at,updated_at FROM "+table+" ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	items := []rule{}
	for rows.Next() {
		var it rule
		if err := rows.Scan(&it.ID, &it.Pattern, &it.RuleType, &it.Comment, &it.Enabled, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}

func (h *Handler) createRule(w http.ResponseWriter, r *http.Request, table string) {
	c := currentUser(r)
	var req struct {
		Pattern  string  `json:"pattern"`
		RuleType string  `json:"rule_type"`
		Comment  *string `json:"comment"`
		Enabled  *int    `json:"enabled"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Pattern == "" {
		respondError(w, http.StatusBadRequest, "pattern is required")
		return
	}
	if req.RuleType == "" {
		req.RuleType = "exact"
	}
	enabled := 1
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	id := uuid.New().String()
	if _, err := h.db.Exec("INSERT INTO "+table+"(id,pattern,rule_type,comment,enabled,created_by) VALUES(?,?,?,?,?,?)",
		id, req.Pattern, req.RuleType, req.Comment, enabled, c.UserID); err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	var it rule
	_ = h.db.QueryRow("SELECT id,pattern,rule_type,comment,enabled,created_by,created_at,updated_at FROM "+table+" WHERE id=?", id).
		Scan(&it.ID, &it.Pattern, &it.RuleType, &it.Comment, &it.Enabled, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusCreated, it)
}

func (h *Handler) getRule(w http.ResponseWriter, r *http.Request, table string) {
	id := chi.URLParam(r, "id")
	var it rule
	err := h.db.QueryRow("SELECT id,pattern,rule_type,comment,enabled,created_by,created_at,updated_at FROM "+table+" WHERE id=?", id).
		Scan(&it.ID, &it.Pattern, &it.RuleType, &it.Comment, &it.Enabled, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respond(w, http.StatusOK, it)
}

func (h *Handler) updateRule(w http.ResponseWriter, r *http.Request, table string) {
	id := chi.URLParam(r, "id")
	var req struct {
		Pattern  *string `json:"pattern"`
		RuleType *string `json:"rule_type"`
		Comment  *string `json:"comment"`
		Enabled  *int    `json:"enabled"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Pattern != nil && *req.Pattern == "" {
		respondError(w, http.StatusBadRequest, "pattern cannot be empty")
		return
	}
	res, err := h.db.Exec(`UPDATE `+table+` SET
		pattern=COALESCE(?,pattern),
		rule_type=COALESCE(?,rule_type),
		comment=COALESCE(?,comment),
		enabled=COALESCE(?,enabled),
		updated_at=datetime('now')
		WHERE id=?`, req.Pattern, req.RuleType, req.Comment, req.Enabled, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	var it rule
	_ = h.db.QueryRow("SELECT id,pattern,rule_type,comment,enabled,created_by,created_at,updated_at FROM "+table+" WHERE id=?", id).
		Scan(&it.ID, &it.Pattern, &it.RuleType, &it.Comment, &it.Enabled, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusOK, it)
}

func (h *Handler) deleteRule(w http.ResponseWriter, r *http.Request, table string) {
	id := chi.URLParam(r, "id")
	res, err := h.db.Exec("DELETE FROM "+table+" WHERE id=?", id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respond(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// Block rules
func (h *Handler) ListBlockRules(w http.ResponseWriter, r *http.Request) {
	h.listRules(w, r, "block_rules")
}
func (h *Handler) CreateBlockRule(w http.ResponseWriter, r *http.Request) {
	h.createRule(w, r, "block_rules")
}
func (h *Handler) GetBlockRule(w http.ResponseWriter, r *http.Request) {
	h.getRule(w, r, "block_rules")
}
func (h *Handler) UpdateBlockRule(w http.ResponseWriter, r *http.Request) {
	h.updateRule(w, r, "block_rules")
}
func (h *Handler) DeleteBlockRule(w http.ResponseWriter, r *http.Request) {
	h.deleteRule(w, r, "block_rules")
}

// Allow rules
func (h *Handler) ListAllowRules(w http.ResponseWriter, r *http.Request) {
	h.listRules(w, r, "allow_rules")
}
func (h *Handler) CreateAllowRule(w http.ResponseWriter, r *http.Request) {
	h.createRule(w, r, "allow_rules")
}
func (h *Handler) GetAllowRule(w http.ResponseWriter, r *http.Request) {
	h.getRule(w, r, "allow_rules")
}
func (h *Handler) UpdateAllowRule(w http.ResponseWriter, r *http.Request) {
	h.updateRule(w, r, "allow_rules")
}
func (h *Handler) DeleteAllowRule(w http.ResponseWriter, r *http.Request) {
	h.deleteRule(w, r, "allow_rules")
}

// ---- Rule sources ----

type ruleSource struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	URL           string  `json:"url"`
	Format        string  `json:"format"`
	Enabled       int     `json:"enabled"`
	LastUpdatedAt *string `json:"last_updated_at"`
	RuleCount     int     `json:"rule_count"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func (h *Handler) ListRuleSources(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 100)
	offset := queryInt(r, "offset", 0)
	var total int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM rule_sources").Scan(&total)
	rows, err := h.db.Query("SELECT id,name,url,format,enabled,last_updated_at,rule_count,created_at,updated_at FROM rule_sources ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	items := []ruleSource{}
	for rows.Next() {
		var it ruleSource
		if err := rows.Scan(&it.ID, &it.Name, &it.URL, &it.Format, &it.Enabled, &it.LastUpdatedAt, &it.RuleCount, &it.CreatedAt, &it.UpdatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}

func (h *Handler) CreateRuleSource(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		URL     string `json:"url"`
		Format  string `json:"format"`
		Enabled *int   `json:"enabled"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.URL == "" {
		respondError(w, http.StatusBadRequest, "name and url are required")
		return
	}
	if req.Format == "" {
		req.Format = "hosts"
	}
	enabled := 1
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	id := uuid.New().String()
	if _, err := h.db.Exec("INSERT INTO rule_sources(id,name,url,format,enabled) VALUES(?,?,?,?,?)",
		id, req.Name, req.URL, req.Format, enabled); err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	var it ruleSource
	_ = h.db.QueryRow("SELECT id,name,url,format,enabled,last_updated_at,rule_count,created_at,updated_at FROM rule_sources WHERE id=?", id).
		Scan(&it.ID, &it.Name, &it.URL, &it.Format, &it.Enabled, &it.LastUpdatedAt, &it.RuleCount, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusCreated, it)
}

func (h *Handler) GetRuleSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var it ruleSource
	err := h.db.QueryRow("SELECT id,name,url,format,enabled,last_updated_at,rule_count,created_at,updated_at FROM rule_sources WHERE id=?", id).
		Scan(&it.ID, &it.Name, &it.URL, &it.Format, &it.Enabled, &it.LastUpdatedAt, &it.RuleCount, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respond(w, http.StatusOK, it)
}

func (h *Handler) UpdateRuleSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name    *string `json:"name"`
		URL     *string `json:"url"`
		Format  *string `json:"format"`
		Enabled *int    `json:"enabled"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	res, err := h.db.Exec(`UPDATE rule_sources SET
		name=COALESCE(?,name),
		url=COALESCE(?,url),
		format=COALESCE(?,format),
		enabled=COALESCE(?,enabled),
		updated_at=datetime('now')
		WHERE id=?`, req.Name, req.URL, req.Format, req.Enabled, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	var it ruleSource
	_ = h.db.QueryRow("SELECT id,name,url,format,enabled,last_updated_at,rule_count,created_at,updated_at FROM rule_sources WHERE id=?", id).
		Scan(&it.ID, &it.Name, &it.URL, &it.Format, &it.Enabled, &it.LastUpdatedAt, &it.RuleCount, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusOK, it)
}

func (h *Handler) DeleteRuleSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.db.Exec("DELETE FROM rule_sources WHERE id=?", id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respond(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *Handler) RefreshRuleSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	res, err := h.db.Exec("UPDATE rule_sources SET last_updated_at=datetime('now'),updated_at=datetime('now') WHERE id=?", id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respond(w, http.StatusOK, map[string]string{"message": "refreshed"})
}
