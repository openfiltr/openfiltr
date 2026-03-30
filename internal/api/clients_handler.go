package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/openfiltr/openfiltr/internal/storage"
)

// ---- Clients ----

type client = storage.ClientView

func (h *Handler) ListClients(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 100)
	offset := queryInt(r, "offset", 0)
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		items, total, err := bolt.ListClients(limit, offset)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
		return
	}
	var total int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM clients").Scan(&total)
	rows, err := h.db.Query(storage.Rebind("SELECT id,name,identifier,identifier_type,group_id,comment,created_at::text,updated_at::text FROM clients ORDER BY created_at DESC LIMIT ? OFFSET ?"), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	items := []client{}
	for rows.Next() {
		var it client
		if err := rows.Scan(&it.ID, &it.Name, &it.Identifier, &it.IdentifierType, &it.GroupID, &it.Comment, &it.CreatedAt, &it.UpdatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}

func (h *Handler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name           string  `json:"name"`
		Identifier     string  `json:"identifier"`
		IdentifierType string  `json:"identifier_type"`
		GroupID        *string `json:"group_id"`
		Comment        *string `json:"comment"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Identifier == "" {
		respondError(w, http.StatusBadRequest, "name and identifier are required")
		return
	}
	if req.IdentifierType == "" {
		req.IdentifierType = "ip"
	}
	id := uuid.New().String()
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, err := bolt.CreateClient(id, req.Name, req.Identifier, req.IdentifierType, req.GroupID, req.Comment)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		respond(w, http.StatusCreated, it)
		return
	}
	if _, err := h.db.Exec(storage.Rebind("INSERT INTO clients(id,name,identifier,identifier_type,group_id,comment) VALUES(?,?,?,?,?,?)"),
		id, req.Name, req.Identifier, req.IdentifierType, req.GroupID, req.Comment); err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	var it client
	_ = h.db.QueryRow(storage.Rebind("SELECT id,name,identifier,identifier_type,group_id,comment,created_at::text,updated_at::text FROM clients WHERE id=?"), id).
		Scan(&it.ID, &it.Name, &it.Identifier, &it.IdentifierType, &it.GroupID, &it.Comment, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusCreated, it)
}

func (h *Handler) GetClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, err := bolt.GetClient(id)
		if err != nil {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respond(w, http.StatusOK, it)
		return
	}
	var it client
	err := h.db.QueryRow(storage.Rebind("SELECT id,name,identifier,identifier_type,group_id,comment,created_at::text,updated_at::text FROM clients WHERE id=?"), id).
		Scan(&it.ID, &it.Name, &it.Identifier, &it.IdentifierType, &it.GroupID, &it.Comment, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respond(w, http.StatusOK, it)
}

func (h *Handler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name           *string `json:"name"`
		Identifier     *string `json:"identifier"`
		IdentifierType *string `json:"identifier_type"`
		GroupID        *string `json:"group_id"`
		Comment        *string `json:"comment"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, found, err := bolt.UpdateClient(id, req.Name, req.Identifier, req.IdentifierType, req.GroupID, req.Comment)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		if !found {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respond(w, http.StatusOK, it)
		return
	}
	res, err := h.db.Exec(storage.Rebind(`UPDATE clients SET
		name=COALESCE(?,name),
		identifier=COALESCE(?,identifier),
		identifier_type=COALESCE(?,identifier_type),
		group_id=COALESCE(?,group_id),
		comment=COALESCE(?,comment),
		updated_at=NOW()
		WHERE id=?`), req.Name, req.Identifier, req.IdentifierType, req.GroupID, req.Comment, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	var it client
	_ = h.db.QueryRow(storage.Rebind("SELECT id,name,identifier,identifier_type,group_id,comment,created_at::text,updated_at::text FROM clients WHERE id=?"), id).
		Scan(&it.ID, &it.Name, &it.Identifier, &it.IdentifierType, &it.GroupID, &it.Comment, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusOK, it)
}

func (h *Handler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		deleted, err := bolt.DeleteClient(id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		if !deleted {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respond(w, http.StatusOK, map[string]string{"message": "deleted"})
		return
	}
	res, err := h.db.Exec(storage.Rebind("DELETE FROM clients WHERE id=?"), id)
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

// ---- Groups ----

type group = storage.GroupView

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 100)
	offset := queryInt(r, "offset", 0)
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		items, total, err := bolt.ListGroups(limit, offset)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
		return
	}
	var total int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM groups").Scan(&total)
	rows, err := h.db.Query(storage.Rebind("SELECT id,name,description,created_at::text,updated_at::text FROM groups ORDER BY name ASC LIMIT ? OFFSET ?"), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	items := []group{}
	for rows.Next() {
		var it group
		if err := rows.Scan(&it.ID, &it.Name, &it.Description, &it.CreatedAt, &it.UpdatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	id := uuid.New().String()
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, err := bolt.CreateGroup(id, req.Name, req.Description)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		respond(w, http.StatusCreated, it)
		return
	}
	if _, err := h.db.Exec(storage.Rebind("INSERT INTO groups(id,name,description) VALUES(?,?,?)"),
		id, req.Name, req.Description); err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	var it group
	_ = h.db.QueryRow(storage.Rebind("SELECT id,name,description,created_at::text,updated_at::text FROM groups WHERE id=?"), id).
		Scan(&it.ID, &it.Name, &it.Description, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusCreated, it)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, err := bolt.GetGroup(id)
		if err != nil {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respond(w, http.StatusOK, it)
		return
	}
	var it group
	err := h.db.QueryRow(storage.Rebind("SELECT id,name,description,created_at::text,updated_at::text FROM groups WHERE id=?"), id).
		Scan(&it.ID, &it.Name, &it.Description, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respond(w, http.StatusOK, it)
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, found, err := bolt.UpdateGroup(id, req.Name, req.Description)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		if !found {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respond(w, http.StatusOK, it)
		return
	}
	res, err := h.db.Exec(storage.Rebind(`UPDATE groups SET
		name=COALESCE(?,name),
		description=COALESCE(?,description),
		updated_at=NOW()
		WHERE id=?`), req.Name, req.Description, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	var it group
	_ = h.db.QueryRow(storage.Rebind("SELECT id,name,description,created_at::text,updated_at::text FROM groups WHERE id=?"), id).
		Scan(&it.ID, &it.Name, &it.Description, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusOK, it)
}

func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		deleted, err := bolt.DeleteGroup(id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		if !deleted {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respond(w, http.StatusOK, map[string]string{"message": "deleted"})
		return
	}
	res, err := h.db.Exec(storage.Rebind("DELETE FROM groups WHERE id=?"), id)
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
