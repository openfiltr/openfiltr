package api

import (
	"net/http"

	"github.com/openfiltr/openfiltr/internal/storage"
)

type activityEntry struct {
	ID             string  `json:"id"`
	ClientIP       string  `json:"client_ip"`
	Domain         string  `json:"domain"`
	QueryType      string  `json:"query_type"`
	Action         string  `json:"action"`
	RuleID         *string `json:"rule_id"`
	RuleSource     *string `json:"rule_source"`
	ResponseTimeMs *int    `json:"response_time_ms"`
	CreatedAt      string  `json:"created_at"`
}

func (h *Handler) ListActivity(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 100)
	offset := queryInt(r, "offset", 0)

	clientIP := r.URL.Query().Get("client_ip")
	domain := r.URL.Query().Get("domain")
	action := r.URL.Query().Get("action")

	where := " WHERE 1=1"
	args := []interface{}{}
	if clientIP != "" {
		where += " AND client_ip=?"
		args = append(args, clientIP)
	}
	if domain != "" {
		where += " AND domain LIKE ?"
		args = append(args, "%"+domain+"%")
	}
	if action != "" {
		where += " AND action=?"
		args = append(args, action)
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	_ = h.db.QueryRow(storage.Rebind("SELECT COUNT(*) FROM activity_log"+where), countArgs...).Scan(&total)

	args = append(args, limit, offset)
	rows, err := h.db.Query(storage.Rebind("SELECT id,client_ip,domain,query_type,action,rule_id,rule_source,response_time_ms,created_at::text FROM activity_log"+where+" ORDER BY created_at DESC LIMIT ? OFFSET ?"), args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	items := []activityEntry{}
	for rows.Next() {
		var it activityEntry
		if err := rows.Scan(&it.ID, &it.ClientIP, &it.Domain, &it.QueryType, &it.Action, &it.RuleID, &it.RuleSource, &it.ResponseTimeMs, &it.CreatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	respond(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}

func (h *Handler) ActivityStats(w http.ResponseWriter, r *http.Request) {
	var total, blocked, allowed int
	_ = h.db.QueryRow("SELECT COUNT(*) FROM activity_log").Scan(&total)
	_ = h.db.QueryRow("SELECT COUNT(*) FROM activity_log WHERE action='blocked'").Scan(&blocked)
	_ = h.db.QueryRow("SELECT COUNT(*) FROM activity_log WHERE action='allowed'").Scan(&allowed)

	type topDomain struct {
		Domain string `json:"domain"`
		Count  int    `json:"count"`
	}
	rows, err := h.db.Query(`SELECT domain,COUNT(*) as cnt FROM activity_log WHERE action='blocked' GROUP BY domain ORDER BY cnt DESC LIMIT 10`)
	topBlocked := []topDomain{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var td topDomain
			if err := rows.Scan(&td.Domain, &td.Count); err == nil {
				topBlocked = append(topBlocked, td)
			}
		}
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"total":               total,
		"blocked":             blocked,
		"allowed":             allowed,
		"top_blocked_domains": topBlocked,
	})
}
