package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/openfiltr/openfiltr/internal/storage"
)

// ---- Upstream servers ----

type upstreamServer = storage.UpstreamServerView

func (h *Handler) ListUpstreamServers(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 100)
	offset := queryInt(r, "offset", 0)
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		items, total, err := bolt.ListUpstreamServers(limit, offset)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
		return
	}
	var total int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM upstream_servers").Scan(&total)
	rows, err := h.db.Query(storage.Rebind("SELECT id,name,address,protocol,enabled,priority,created_at::text,updated_at::text FROM upstream_servers ORDER BY priority ASC,created_at DESC LIMIT ? OFFSET ?"), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	items := []upstreamServer{}
	for rows.Next() {
		var it upstreamServer
		if err := rows.Scan(&it.ID, &it.Name, &it.Address, &it.Protocol, &it.Enabled, &it.Priority, &it.CreatedAt, &it.UpdatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}

func (h *Handler) CreateUpstreamServer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Address  string `json:"address"`
		Protocol string `json:"protocol"`
		Enabled  *int   `json:"enabled"`
		Priority *int   `json:"priority"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Address == "" {
		respondError(w, http.StatusBadRequest, "name and address are required")
		return
	}
	if req.Protocol == "" {
		req.Protocol = "udp"
	}
	enabled := 1
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	priority := 0
	if req.Priority != nil {
		priority = *req.Priority
	}
	id := uuid.New().String()
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, err := bolt.CreateUpstreamServer(id, req.Name, req.Address, req.Protocol, enabled, priority)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		respond(w, http.StatusCreated, it)
		return
	}
	if _, err := h.db.Exec(storage.Rebind("INSERT INTO upstream_servers(id,name,address,protocol,enabled,priority) VALUES(?,?,?,?,?,?)"),
		id, req.Name, req.Address, req.Protocol, enabled, priority); err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	var it upstreamServer
	_ = h.db.QueryRow(storage.Rebind("SELECT id,name,address,protocol,enabled,priority,created_at::text,updated_at::text FROM upstream_servers WHERE id=?"), id).
		Scan(&it.ID, &it.Name, &it.Address, &it.Protocol, &it.Enabled, &it.Priority, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusCreated, it)
}

func (h *Handler) GetUpstreamServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, err := bolt.GetUpstreamServer(id)
		if err != nil {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respond(w, http.StatusOK, it)
		return
	}
	var it upstreamServer
	err := h.db.QueryRow(storage.Rebind("SELECT id,name,address,protocol,enabled,priority,created_at::text,updated_at::text FROM upstream_servers WHERE id=?"), id).
		Scan(&it.ID, &it.Name, &it.Address, &it.Protocol, &it.Enabled, &it.Priority, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respond(w, http.StatusOK, it)
}

func (h *Handler) UpdateUpstreamServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name     *string `json:"name"`
		Address  *string `json:"address"`
		Protocol *string `json:"protocol"`
		Enabled  *int    `json:"enabled"`
		Priority *int    `json:"priority"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, found, err := bolt.UpdateUpstreamServer(id, req.Name, req.Address, req.Protocol, req.Enabled, req.Priority)
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
	res, err := h.db.Exec(storage.Rebind(`UPDATE upstream_servers SET
		name=COALESCE(?,name),
		address=COALESCE(?,address),
		protocol=COALESCE(?,protocol),
		enabled=COALESCE(?,enabled),
		priority=COALESCE(?,priority),
		updated_at=NOW()
		WHERE id=?`), req.Name, req.Address, req.Protocol, req.Enabled, req.Priority, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	var it upstreamServer
	_ = h.db.QueryRow(storage.Rebind("SELECT id,name,address,protocol,enabled,priority,created_at::text,updated_at::text FROM upstream_servers WHERE id=?"), id).
		Scan(&it.ID, &it.Name, &it.Address, &it.Protocol, &it.Enabled, &it.Priority, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusOK, it)
}

func (h *Handler) DeleteUpstreamServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		deleted, err := bolt.DeleteUpstreamServer(id)
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
	res, err := h.db.Exec(storage.Rebind("DELETE FROM upstream_servers WHERE id=?"), id)
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

// ---- DNS entries ----

type dnsEntry = storage.DNSEntryView

func (h *Handler) ListDNSEntries(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 100)
	offset := queryInt(r, "offset", 0)
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		items, total, err := bolt.ListDNSEntries(limit, offset)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
		return
	}
	var total int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM dns_entries").Scan(&total)
	rows, err := h.db.Query(storage.Rebind("SELECT id,host,entry_type,value,ttl,comment,enabled,created_by,created_at::text,updated_at::text FROM dns_entries ORDER BY created_at DESC LIMIT ? OFFSET ?"), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	items := []dnsEntry{}
	for rows.Next() {
		var it dnsEntry
		if err := rows.Scan(&it.ID, &it.Host, &it.EntryType, &it.Value, &it.TTL, &it.Comment, &it.Enabled, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}

func (h *Handler) CreateDNSEntry(w http.ResponseWriter, r *http.Request) {
	c := currentUser(r)
	var req struct {
		Host      string  `json:"host"`
		EntryType string  `json:"entry_type"`
		Value     string  `json:"value"`
		TTL       *int    `json:"ttl"`
		Comment   *string `json:"comment"`
		Enabled   *int    `json:"enabled"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Host == "" || req.EntryType == "" || req.Value == "" {
		respondError(w, http.StatusBadRequest, "host, entry_type and value are required")
		return
	}
	ttl := 300
	if req.TTL != nil {
		ttl = *req.TTL
	}
	enabled := 1
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	id := uuid.New().String()
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		createdBy := c.UserID
		it, err := bolt.CreateDNSEntry(id, req.Host, req.EntryType, req.Value, ttl, req.Comment, &createdBy, enabled)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		respond(w, http.StatusCreated, it)
		return
	}
	if _, err := h.db.Exec(storage.Rebind("INSERT INTO dns_entries(id,host,entry_type,value,ttl,comment,enabled,created_by) VALUES(?,?,?,?,?,?,?,?)"),
		id, req.Host, req.EntryType, req.Value, ttl, req.Comment, enabled, c.UserID); err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	var it dnsEntry
	_ = h.db.QueryRow(storage.Rebind("SELECT id,host,entry_type,value,ttl,comment,enabled,created_by,created_at::text,updated_at::text FROM dns_entries WHERE id=?"), id).
		Scan(&it.ID, &it.Host, &it.EntryType, &it.Value, &it.TTL, &it.Comment, &it.Enabled, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusCreated, it)
}

func (h *Handler) GetDNSEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, err := bolt.GetDNSEntry(id)
		if err != nil {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respond(w, http.StatusOK, it)
		return
	}
	var it dnsEntry
	err := h.db.QueryRow(storage.Rebind("SELECT id,host,entry_type,value,ttl,comment,enabled,created_by,created_at::text,updated_at::text FROM dns_entries WHERE id=?"), id).
		Scan(&it.ID, &it.Host, &it.EntryType, &it.Value, &it.TTL, &it.Comment, &it.Enabled, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respond(w, http.StatusOK, it)
}

func (h *Handler) UpdateDNSEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Host      *string `json:"host"`
		EntryType *string `json:"entry_type"`
		Value     *string `json:"value"`
		TTL       *int    `json:"ttl"`
		Comment   *string `json:"comment"`
		Enabled   *int    `json:"enabled"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		it, found, err := bolt.UpdateDNSEntry(id, req.Host, req.EntryType, req.Value, req.TTL, req.Comment, req.Enabled)
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
	res, err := h.db.Exec(storage.Rebind(`UPDATE dns_entries SET
		host=COALESCE(?,host),
		entry_type=COALESCE(?,entry_type),
		value=COALESCE(?,value),
		ttl=COALESCE(?,ttl),
		comment=COALESCE(?,comment),
		enabled=COALESCE(?,enabled),
		updated_at=NOW()
		WHERE id=?`), req.Host, req.EntryType, req.Value, req.TTL, req.Comment, req.Enabled, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	var it dnsEntry
	_ = h.db.QueryRow(storage.Rebind("SELECT id,host,entry_type,value,ttl,comment,enabled,created_by,created_at::text,updated_at::text FROM dns_entries WHERE id=?"), id).
		Scan(&it.ID, &it.Host, &it.EntryType, &it.Value, &it.TTL, &it.Comment, &it.Enabled, &it.CreatedBy, &it.CreatedAt, &it.UpdatedAt)
	respond(w, http.StatusOK, it)
}

func (h *Handler) DeleteDNSEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		deleted, err := bolt.DeleteDNSEntry(id)
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
	res, err := h.db.Exec(storage.Rebind("DELETE FROM dns_entries WHERE id=?"), id)
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
