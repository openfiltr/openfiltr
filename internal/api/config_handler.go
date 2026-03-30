package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/openfiltr/openfiltr/internal/storage"
	"gopkg.in/yaml.v3"
)

const configExportVersion = 1

type exportPayload struct {
	Version         int                      `yaml:"version"`
	BlockRules      []map[string]interface{} `yaml:"block_rules"`
	AllowRules      []map[string]interface{} `yaml:"allow_rules"`
	RuleSources     []map[string]interface{} `yaml:"rule_sources"`
	DNSEntries      []map[string]interface{} `yaml:"dns_entries"`
	UpstreamServers []map[string]interface{} `yaml:"upstream_servers"`
}

func (h *Handler) ExportConfig(w http.ResponseWriter, r *http.Request) {
	payload := exportPayload{Version: configExportVersion}
	if bolt, ok := h.db.(*storage.BoltStore); ok {
		var err error
		payload, err = exportConfigFromBolt(bolt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to load config")
			return
		}
	} else {
		payload.BlockRules = h.fetchRows("SELECT id,pattern,rule_type,comment,enabled FROM block_rules")
		payload.AllowRules = h.fetchRows("SELECT id,pattern,rule_type,comment,enabled FROM allow_rules")
		payload.RuleSources = h.fetchRows("SELECT id,name,url,format,enabled FROM rule_sources")
		payload.DNSEntries = h.fetchRows("SELECT id,host,entry_type,value,ttl,comment,enabled FROM dns_entries")
		payload.UpstreamServers = h.fetchRows("SELECT id,name,address,protocol,enabled,priority FROM upstream_servers")
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
			row[col] = normaliseRowValue(vals[i])
		}
		result = append(result, row)
	}
	return result
}

func (h *Handler) ImportConfig(w http.ResponseWriter, r *http.Request) {
	var payload exportPayload
	if err := yaml.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid config payload: "+err.Error())
		return
	}
	if err := validateConfigVersion(payload.Version); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if bolt, ok := h.db.(*storage.BoltStore); ok {
		imported, err := importConfigToBolt(bolt, payload)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to import config")
			return
		}
		respond(w, http.StatusOK, map[string]interface{}{"imported": imported})
		return
	}

	imported := 0

	for _, row := range payload.BlockRules {
		id := stringValue(row, "id")
		pattern := stringValue(row, "pattern")
		ruleType := stringValue(row, "rule_type")
		comment := nullableStringValue(row, "comment")
		enabled := intValue(row, "enabled", 1)
		if pattern == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		if ruleType == "" {
			ruleType = "exact"
		}
		_, _ = h.db.Exec(storage.Rebind(`INSERT INTO block_rules(id,pattern,rule_type,comment,enabled) VALUES(?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET pattern=excluded.pattern,rule_type=excluded.rule_type,comment=excluded.comment,enabled=excluded.enabled,updated_at=NOW()`),
			id, pattern, ruleType, comment, enabled)
		imported++
	}

	for _, row := range payload.AllowRules {
		id := stringValue(row, "id")
		pattern := stringValue(row, "pattern")
		ruleType := stringValue(row, "rule_type")
		comment := nullableStringValue(row, "comment")
		enabled := intValue(row, "enabled", 1)
		if pattern == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		if ruleType == "" {
			ruleType = "exact"
		}
		_, _ = h.db.Exec(storage.Rebind(`INSERT INTO allow_rules(id,pattern,rule_type,comment,enabled) VALUES(?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET pattern=excluded.pattern,rule_type=excluded.rule_type,comment=excluded.comment,enabled=excluded.enabled,updated_at=NOW()`),
			id, pattern, ruleType, comment, enabled)
		imported++
	}

	for _, row := range payload.RuleSources {
		id := stringValue(row, "id")
		name := stringValue(row, "name")
		url := stringValue(row, "url")
		format := stringValue(row, "format")
		enabled := intValue(row, "enabled", 1)
		if name == "" || url == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		if format == "" {
			format = "hosts"
		}
		_, _ = h.db.Exec(storage.Rebind(`INSERT INTO rule_sources(id,name,url,format,enabled) VALUES(?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,url=excluded.url,format=excluded.format,enabled=excluded.enabled,updated_at=NOW()`),
			id, name, url, format, enabled)
		imported++
	}

	for _, row := range payload.DNSEntries {
		id := stringValue(row, "id")
		host := stringValue(row, "host")
		entryType := stringValue(row, "entry_type")
		value := stringValue(row, "value")
		comment := nullableStringValue(row, "comment")
		ttl := intValue(row, "ttl", 300)
		enabled := intValue(row, "enabled", 1)
		if host == "" || entryType == "" || value == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		_, _ = h.db.Exec(storage.Rebind(`INSERT INTO dns_entries(id,host,entry_type,value,ttl,comment,enabled) VALUES(?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET host=excluded.host,entry_type=excluded.entry_type,value=excluded.value,ttl=excluded.ttl,comment=excluded.comment,enabled=excluded.enabled,updated_at=NOW()`),
			id, host, entryType, value, ttl, comment, enabled)
		imported++
	}

	for _, row := range payload.UpstreamServers {
		id := stringValue(row, "id")
		name := stringValue(row, "name")
		address := stringValue(row, "address")
		protocol := stringValue(row, "protocol")
		enabled := intValue(row, "enabled", 1)
		priority := intValue(row, "priority", 0)
		if name == "" || address == "" {
			continue
		}
		if id == "" {
			id = newID()
		}
		if protocol == "" {
			protocol = "udp"
		}
		_, _ = h.db.Exec(storage.Rebind(`INSERT INTO upstream_servers(id,name,address,protocol,enabled,priority) VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,address=excluded.address,protocol=excluded.protocol,enabled=excluded.enabled,priority=excluded.priority,updated_at=NOW()`),
			id, name, address, protocol, enabled, priority)
		imported++
	}

	respond(w, http.StatusOK, map[string]interface{}{"imported": imported})
}

func exportConfigFromBolt(bolt *storage.BoltStore) (exportPayload, error) {
	limit := int(^uint(0) >> 1)

	blockRules, _, err := bolt.ListRules("block_rules", limit, 0)
	if err != nil {
		return exportPayload{}, err
	}
	allowRules, _, err := bolt.ListRules("allow_rules", limit, 0)
	if err != nil {
		return exportPayload{}, err
	}
	ruleSources, _, err := bolt.ListRuleSources(limit, 0)
	if err != nil {
		return exportPayload{}, err
	}
	dnsEntries, _, err := bolt.ListDNSEntries(limit, 0)
	if err != nil {
		return exportPayload{}, err
	}
	upstreamServers, _, err := bolt.ListUpstreamServers(limit, 0)
	if err != nil {
		return exportPayload{}, err
	}

	return exportPayload{
		Version:         configExportVersion,
		BlockRules:      exportRuleMaps(blockRules),
		AllowRules:      exportRuleMaps(allowRules),
		RuleSources:     exportRuleSourceMaps(ruleSources),
		DNSEntries:      exportDNSEntryMaps(dnsEntries),
		UpstreamServers: exportUpstreamServerMaps(upstreamServers),
	}, nil
}

func importConfigToBolt(bolt *storage.BoltStore, payload exportPayload) (int, error) {
	imported := 0
	for _, row := range payload.BlockRules {
		if err := importBoltRule(bolt, "block_rules", row); err != nil {
			return 0, err
		}
		if stringValue(row, "pattern") != "" {
			imported++
		}
	}
	for _, row := range payload.AllowRules {
		if err := importBoltRule(bolt, "allow_rules", row); err != nil {
			return 0, err
		}
		if stringValue(row, "pattern") != "" {
			imported++
		}
	}
	for _, row := range payload.RuleSources {
		if err := importBoltRuleSource(bolt, row); err != nil {
			return 0, err
		}
		if stringValue(row, "name") != "" && stringValue(row, "url") != "" {
			imported++
		}
	}
	for _, row := range payload.DNSEntries {
		if err := importBoltDNSEntry(bolt, row); err != nil {
			return 0, err
		}
		if stringValue(row, "host") != "" && stringValue(row, "entry_type") != "" && stringValue(row, "value") != "" {
			imported++
		}
	}
	for _, row := range payload.UpstreamServers {
		if err := importBoltUpstreamServer(bolt, row); err != nil {
			return 0, err
		}
		if stringValue(row, "name") != "" && stringValue(row, "address") != "" {
			imported++
		}
	}
	return imported, nil
}

func importBoltRule(bolt *storage.BoltStore, table string, row map[string]interface{}) error {
	id := stringValue(row, "id")
	pattern := stringValue(row, "pattern")
	ruleType := stringValue(row, "rule_type")
	comment := nullableStringValue(row, "comment")
	enabled := intValue(row, "enabled", 1)
	if pattern == "" {
		return nil
	}
	if id == "" {
		id = newID()
	}
	if ruleType == "" {
		ruleType = "exact"
	}
	patternCopy := pattern
	ruleTypeCopy := ruleType
	enabledCopy := enabled
	if _, found, err := bolt.UpdateRule(table, id, &patternCopy, &ruleTypeCopy, comment, &enabledCopy); err != nil {
		return err
	} else if found {
		return nil
	}
	_, err := bolt.CreateRule(table, id, pattern, ruleType, comment, enabled, nil)
	return err
}

func importBoltRuleSource(bolt *storage.BoltStore, row map[string]interface{}) error {
	id := stringValue(row, "id")
	name := stringValue(row, "name")
	url := stringValue(row, "url")
	format := stringValue(row, "format")
	enabled := intValue(row, "enabled", 1)
	if name == "" || url == "" {
		return nil
	}
	if id == "" {
		id = newID()
	}
	if format == "" {
		format = "hosts"
	}
	nameCopy := name
	urlCopy := url
	formatCopy := format
	enabledCopy := enabled
	if _, found, err := bolt.UpdateRuleSource(id, &nameCopy, &urlCopy, &formatCopy, &enabledCopy); err != nil {
		return err
	} else if found {
		return nil
	}
	_, err := bolt.CreateRuleSource(id, name, url, format, enabled)
	return err
}

func importBoltDNSEntry(bolt *storage.BoltStore, row map[string]interface{}) error {
	id := stringValue(row, "id")
	host := stringValue(row, "host")
	entryType := stringValue(row, "entry_type")
	value := stringValue(row, "value")
	comment := nullableStringValue(row, "comment")
	ttl := intValue(row, "ttl", 300)
	enabled := intValue(row, "enabled", 1)
	if host == "" || entryType == "" || value == "" {
		return nil
	}
	if id == "" {
		id = newID()
	}
	hostCopy := host
	entryTypeCopy := entryType
	valueCopy := value
	ttlCopy := ttl
	enabledCopy := enabled
	if _, found, err := bolt.UpdateDNSEntry(id, &hostCopy, &entryTypeCopy, &valueCopy, &ttlCopy, comment, &enabledCopy); err != nil {
		return err
	} else if found {
		return nil
	}
	_, err := bolt.CreateDNSEntry(id, host, entryType, value, ttl, comment, nil, enabled)
	return err
}

func importBoltUpstreamServer(bolt *storage.BoltStore, row map[string]interface{}) error {
	id := stringValue(row, "id")
	name := stringValue(row, "name")
	address := stringValue(row, "address")
	protocol := stringValue(row, "protocol")
	enabled := intValue(row, "enabled", 1)
	priority := intValue(row, "priority", 0)
	if name == "" || address == "" {
		return nil
	}
	if id == "" {
		id = newID()
	}
	if protocol == "" {
		protocol = "udp"
	}
	nameCopy := name
	addressCopy := address
	protocolCopy := protocol
	enabledCopy := enabled
	priorityCopy := priority
	if _, found, err := bolt.UpdateUpstreamServer(id, &nameCopy, &addressCopy, &protocolCopy, &enabledCopy, &priorityCopy); err != nil {
		return err
	} else if found {
		return nil
	}
	_, err := bolt.CreateUpstreamServer(id, name, address, protocol, enabled, priority)
	return err
}

func exportRuleMaps(items []storage.RuleView) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		result = append(result, map[string]interface{}{
			"id":        it.ID,
			"pattern":   it.Pattern,
			"rule_type": it.RuleType,
			"comment":   it.Comment,
			"enabled":   it.Enabled,
		})
	}
	return result
}

func exportRuleSourceMaps(items []storage.RuleSourceView) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		result = append(result, map[string]interface{}{
			"id":      it.ID,
			"name":    it.Name,
			"url":     it.URL,
			"format":  it.Format,
			"enabled": it.Enabled,
		})
	}
	return result
}

func exportDNSEntryMaps(items []storage.DNSEntryView) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		result = append(result, map[string]interface{}{
			"id":         it.ID,
			"host":       it.Host,
			"entry_type": it.EntryType,
			"value":      it.Value,
			"ttl":        it.TTL,
			"comment":    it.Comment,
			"enabled":    it.Enabled,
		})
	}
	return result
}

func exportUpstreamServerMaps(items []storage.UpstreamServerView) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		result = append(result, map[string]interface{}{
			"id":       it.ID,
			"name":     it.Name,
			"address":  it.Address,
			"protocol": it.Protocol,
			"enabled":  it.Enabled,
			"priority": it.Priority,
		})
	}
	return result
}

func validateConfigVersion(version int) error {
	if version == 0 {
		return fmt.Errorf("missing required top-level version field")
	}
	if version != configExportVersion {
		return fmt.Errorf("unsupported config version %d: expected version %d", version, configExportVersion)
	}
	return nil
}

func newID() string {
	// Use uuid package indirectly via the already-imported helper
	return generateUUID()
}

func normaliseRowValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []byte:
		return string(v)
	default:
		return value
	}
}

func stringValue(row map[string]interface{}, key string) string {
	switch v := row[key].(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func nullableStringValue(row map[string]interface{}, key string) *string {
	switch v := row[key].(type) {
	case nil:
		return nil
	case string:
		return &v
	case []byte:
		s := string(v)
		return &s
	default:
		return nil
	}
}

func intValue(row map[string]interface{}, key string, fallback int) int {
	switch v := row[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	case []byte:
		n, err := strconv.Atoi(string(v))
		if err == nil {
			return n
		}
	}
	return fallback
}
