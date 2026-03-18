package api

import (
	"net/http"

	"gopkg.in/yaml.v3"
)

type exportPayload struct {
	BlockRules      []map[string]interface{} `yaml:"block_rules"`
	AllowRules      []map[string]interface{} `yaml:"allow_rules"`
	RuleSources     []map[string]interface{} `yaml:"rule_sources"`
	DNSEntries      []map[string]interface{} `yaml:"dns_entries"`
	UpstreamServers []map[string]interface{} `yaml:"upstream_servers"`
}

func (h *Handler) ExportConfig(w http.ResponseWriter, r *http.Request) {
	payload := exportPayload{
		BlockRules:      h.fetchRows("SELECT id,pattern,rule_type,comment,enabled FROM block_rules"),
		AllowRules:      h.fetchRows("SELECT id,pattern,rule_type,comment,enabled FROM allow_rules"),
		RuleSources:     h.fetchRows("SELECT id,name,url,format,enabled FROM rule_sources"),
		DNSEntries:      h.fetchRows("SELECT id,host,entry_type,value,ttl,comment,enabled FROM dns_entries"),
		UpstreamServers: h.fetchRows("SELECT id,name,address,protocol,enabled,priority FROM upstream_servers"),
	}
	data, err := yaml.Marshal(payload)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to marshal config")
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Content-Disposition", `attachment; filename="openfiltr-config.yaml"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) fetchRows(query string) []map[string]interface{} {
	rows, err := h.db.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var result []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = vals[i]
		}
		result = append(result, row)
	}
	return result
}

func (h *Handler) ImportConfig(w http.ResponseWriter, r *http.Request) {
	var payload exportPayload
	if err := yaml.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid YAML: "+err.Error())
		return
	}

	imported := 0

	for _, row := range payload.BlockRules {
		id, _ := row["id"].(string)
		pattern, _ := row["pattern"].(string)
		ruleType, _ := row["rule_type"].(string)
		if pattern == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		if ruleType == "" {
			ruleType = "exact"
		}
		_, _ = h.db.Exec(`INSERT INTO block_rules(id,pattern,rule_type) VALUES(?,?,?) ON CONFLICT(id) DO UPDATE SET pattern=excluded.pattern,rule_type=excluded.rule_type,updated_at=datetime('now')`,
			id, pattern, ruleType)
		imported++
	}

	for _, row := range payload.AllowRules {
		id, _ := row["id"].(string)
		pattern, _ := row["pattern"].(string)
		ruleType, _ := row["rule_type"].(string)
		if pattern == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		if ruleType == "" {
			ruleType = "exact"
		}
		_, _ = h.db.Exec(`INSERT INTO allow_rules(id,pattern,rule_type) VALUES(?,?,?) ON CONFLICT(id) DO UPDATE SET pattern=excluded.pattern,rule_type=excluded.rule_type,updated_at=datetime('now')`,
			id, pattern, ruleType)
		imported++
	}

	for _, row := range payload.RuleSources {
		id, _ := row["id"].(string)
		name, _ := row["name"].(string)
		url, _ := row["url"].(string)
		format, _ := row["format"].(string)
		if name == "" || url == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		if format == "" {
			format = "hosts"
		}
		_, _ = h.db.Exec(`INSERT INTO rule_sources(id,name,url,format) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,url=excluded.url,format=excluded.format,updated_at=datetime('now')`,
			id, name, url, format)
		imported++
	}

	for _, row := range payload.DNSEntries {
		id, _ := row["id"].(string)
		host, _ := row["host"].(string)
		entryType, _ := row["entry_type"].(string)
		value, _ := row["value"].(string)
		if host == "" || entryType == "" || value == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		_, _ = h.db.Exec(`INSERT INTO dns_entries(id,host,entry_type,value) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET host=excluded.host,entry_type=excluded.entry_type,value=excluded.value,updated_at=datetime('now')`,
			id, host, entryType, value)
		imported++
	}

	for _, row := range payload.UpstreamServers {
		id, _ := row["id"].(string)
		name, _ := row["name"].(string)
		address, _ := row["address"].(string)
		protocol, _ := row["protocol"].(string)
		if name == "" || address == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		if protocol == "" {
			protocol = "udp"
		}
		_, _ = h.db.Exec(`INSERT INTO upstream_servers(id,name,address,protocol) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,address=excluded.address,protocol=excluded.protocol,updated_at=datetime('now')`,
			id, name, address, protocol)
		imported++
	}

	respond(w, http.StatusOK, map[string]interface{}{"imported": imported})
}

func newID() string {
	// Use uuid package indirectly via the already-imported helper
	return generateUUID()
}
