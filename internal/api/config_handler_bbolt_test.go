package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/openfiltr/openfiltr/internal/storage"
	"gopkg.in/yaml.v3"
)

func TestBoltExportAndImportConfigRoundTrip(t *testing.T) {
	source := openBoltConfigHandlerStore(t)
	defer source.Close()

	blockComment := "block comment"
	if _, err := source.CreateRule("block_rules", "block-1", "ads.example.com", "exact", &blockComment, 1, nil); err != nil {
		t.Fatalf("CreateRule() block error = %v", err)
	}
	allowComment := "allow comment"
	if _, err := source.CreateRule("allow_rules", "allow-1", "allowed.example.com", "wildcard", &allowComment, 1, nil); err != nil {
		t.Fatalf("CreateRule() allow error = %v", err)
	}
	if _, err := source.CreateRuleSource("source-1", "hosts list", "https://example.com/hosts", "hosts", 1); err != nil {
		t.Fatalf("CreateRuleSource() error = %v", err)
	}
	if _, err := source.CreateDNSEntry("dns-1", "example.com", "A", "1.1.1.1", 60, nil, nil, 1); err != nil {
		t.Fatalf("CreateDNSEntry() error = %v", err)
	}
	if _, err := source.CreateUpstreamServer("upstream-1", "primary", "1.1.1.1:53", "udp", 1, 10); err != nil {
		t.Fatalf("CreateUpstreamServer() error = %v", err)
	}

	exportHandler := &Handler{db: source}
	exportRec := httptest.NewRecorder()
	exportHandler.ExportConfig(exportRec, httptest.NewRequest(http.MethodGet, "/api/v1/config/export", nil))
	if exportRec.Code != http.StatusOK {
		t.Fatalf("ExportConfig() status = %d, want %d", exportRec.Code, http.StatusOK)
	}

	var exported exportPayload
	if err := yaml.Unmarshal(exportRec.Body.Bytes(), &exported); err != nil {
		t.Fatalf("yaml.Unmarshal() exported payload = %v", err)
	}
	if exported.Version != configExportVersion {
		t.Fatalf("exported version = %d, want %d", exported.Version, configExportVersion)
	}
	if len(exported.BlockRules) != 1 || stringValue(exported.BlockRules[0], "pattern") != "ads.example.com" {
		t.Fatalf("exported block rules = %#v", exported.BlockRules)
	}
	if len(exported.AllowRules) != 1 || stringValue(exported.AllowRules[0], "rule_type") != "wildcard" {
		t.Fatalf("exported allow rules = %#v", exported.AllowRules)
	}
	if len(exported.RuleSources) != 1 || stringValue(exported.RuleSources[0], "name") != "hosts list" {
		t.Fatalf("exported rule sources = %#v", exported.RuleSources)
	}
	if len(exported.DNSEntries) != 1 || stringValue(exported.DNSEntries[0], "host") != "example.com" {
		t.Fatalf("exported DNS entries = %#v", exported.DNSEntries)
	}
	if len(exported.UpstreamServers) != 1 || stringValue(exported.UpstreamServers[0], "name") != "primary" {
		t.Fatalf("exported upstream servers = %#v", exported.UpstreamServers)
	}

	dest := openBoltConfigHandlerStore(t)
	defer dest.Close()

	importHandler := &Handler{db: dest}
	importRec := httptest.NewRecorder()
	importHandler.ImportConfig(importRec, httptest.NewRequest(http.MethodPost, "/api/v1/config/import", bytes.NewReader(exportRec.Body.Bytes())))
	if importRec.Code != http.StatusOK {
		t.Fatalf("ImportConfig() status = %d, want %d", importRec.Code, http.StatusOK)
	}
	var imported struct {
		Imported int `json:"imported"`
	}
	if err := json.NewDecoder(importRec.Body).Decode(&imported); err != nil {
		t.Fatalf("decoding ImportConfig() response: %v", err)
	}
	if imported.Imported != 5 {
		t.Fatalf("ImportConfig() imported = %d, want %d", imported.Imported, 5)
	}

	gotRule, err := dest.GetRule("block_rules", "block-1")
	if err != nil {
		t.Fatalf("GetRule() error = %v", err)
	}
	if gotRule.Pattern != "ads.example.com" || gotRule.RuleType != "exact" {
		t.Fatalf("GetRule() = %#v, want imported block rule", gotRule)
	}

	gotAllow, err := dest.GetRule("allow_rules", "allow-1")
	if err != nil {
		t.Fatalf("GetRule() allow error = %v", err)
	}
	if gotAllow.Pattern != "allowed.example.com" || gotAllow.RuleType != "wildcard" {
		t.Fatalf("GetRule() allow = %#v, want imported allow rule", gotAllow)
	}

	gotSource, err := dest.GetRuleSource("source-1")
	if err != nil {
		t.Fatalf("GetRuleSource() error = %v", err)
	}
	if gotSource.Name != "hosts list" || gotSource.URL != "https://example.com/hosts" {
		t.Fatalf("GetRuleSource() = %#v, want imported rule source", gotSource)
	}

	gotDNS, err := dest.GetDNSEntry("dns-1")
	if err != nil {
		t.Fatalf("GetDNSEntry() error = %v", err)
	}
	if gotDNS.Host != "example.com" || gotDNS.Value != "1.1.1.1" || gotDNS.TTL != 60 {
		t.Fatalf("GetDNSEntry() = %#v, want imported DNS entry", gotDNS)
	}

	gotUpstream, err := dest.GetUpstreamServer("upstream-1")
	if err != nil {
		t.Fatalf("GetUpstreamServer() error = %v", err)
	}
	if gotUpstream.Name != "primary" || gotUpstream.Address != "1.1.1.1:53" || gotUpstream.Priority != 10 {
		t.Fatalf("GetUpstreamServer() = %#v, want imported upstream server", gotUpstream)
	}
}

func openBoltConfigHandlerStore(t *testing.T) *storage.BoltStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openfiltr.db")
	store, err := storage.OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() error = %v", err)
	}
	return store
}
