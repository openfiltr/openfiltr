package api

import (
	"net/http"

	"github.com/openfiltr/openfiltr/internal/storage"
)

type auditEvent = storage.AuditEventView

func (h *Handler) ListAuditEvents(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 100)
	offset := queryInt(r, "offset", 0)
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		items, total, err := bolt.ListAuditEvents(limit, offset)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
		return
	}
	var total int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM audit_events").Scan(&total)
	rows, err := h.db.Query(storage.Rebind("SELECT id,user_id,action,resource_type,resource_id,details,ip_address,created_at::text FROM audit_events ORDER BY created_at DESC LIMIT ? OFFSET ?"), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	items := []auditEvent{}
	for rows.Next() {
		var it auditEvent
		if err := rows.Scan(&it.ID, &it.UserID, &it.Action, &it.ResourceType, &it.ResourceID, &it.Details, &it.IPAddress, &it.CreatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}
