package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/openfiltr/openfiltr/internal/storage"
)

func TestBoltActivityHandlersUseInMemoryCounts(t *testing.T) {
	store := openBoltActivityHandlerStore(t)
	defer store.Close()

	handler := &Handler{db: store}

	responseTime := 15
	ruleID := "rule-1"
	ruleSource := "block-list"
	if err := store.AppendActivityEntry("activity-1", "10.0.0.1", "api.example.com", "A", "blocked", &ruleID, &ruleSource, &responseTime); err != nil {
		t.Fatalf("AppendActivityEntry() first error = %v", err)
	}
	if err := store.AppendActivityEntry("activity-2", "10.0.0.1", "api.example.com", "A", "blocked", &ruleID, &ruleSource, &responseTime); err != nil {
		t.Fatalf("AppendActivityEntry() second error = %v", err)
	}
	if err := store.AppendActivityEntry("activity-3", "10.0.0.2", "www.example.com", "A", "allowed", nil, nil, &responseTime); err != nil {
		t.Fatalf("AppendActivityEntry() third error = %v", err)
	}

	statsRec := httptest.NewRecorder()
	handler.Stats(statsRec, httptest.NewRequest(http.MethodGet, "/api/v1/system/stats", nil))
	if statsRec.Code != http.StatusOK {
		t.Fatalf("Stats() status = %d, want %d", statsRec.Code, http.StatusOK)
	}
	var stats struct {
		TotalQueries   int    `json:"total_queries"`
		BlockedQueries int    `json:"blocked_queries"`
		AllowedQueries int    `json:"allowed_queries"`
		BlockRate      string `json:"block_rate"`
	}
	if err := json.NewDecoder(statsRec.Body).Decode(&stats); err != nil {
		t.Fatalf("decoding Stats() response: %v", err)
	}
	if stats.TotalQueries != 3 || stats.BlockedQueries != 2 || stats.AllowedQueries != 1 {
		t.Fatalf("Stats() = %+v, want counts (3,2,1)", stats)
	}

	activityRec := httptest.NewRecorder()
	handler.ActivityStats(activityRec, httptest.NewRequest(http.MethodGet, "/api/v1/activity/stats", nil))
	if activityRec.Code != http.StatusOK {
		t.Fatalf("ActivityStats() status = %d, want %d", activityRec.Code, http.StatusOK)
	}
	var activityStats struct {
		Total             int `json:"total"`
		Blocked           int `json:"blocked"`
		Allowed           int `json:"allowed"`
		TopBlockedDomains []struct {
			Domain string `json:"domain"`
			Count  int    `json:"count"`
		} `json:"top_blocked_domains"`
	}
	if err := json.NewDecoder(activityRec.Body).Decode(&activityStats); err != nil {
		t.Fatalf("decoding ActivityStats() response: %v", err)
	}
	if activityStats.Total != 3 || activityStats.Blocked != 2 || activityStats.Allowed != 1 {
		t.Fatalf("ActivityStats() = %+v, want counts (3,2,1)", activityStats)
	}
	if len(activityStats.TopBlockedDomains) != 1 || activityStats.TopBlockedDomains[0].Domain != "api.example.com" || activityStats.TopBlockedDomains[0].Count != 2 {
		t.Fatalf("ActivityStats() top blocked domains = %+v, want api.example.com/2", activityStats.TopBlockedDomains)
	}
}

func TestBoltListActivityFiltersAndPaginates(t *testing.T) {
	store := openBoltActivityHandlerStore(t)
	defer store.Close()

	handler := &Handler{db: store}

	responseTime := 15
	ruleID := "rule-1"
	ruleSource := "block-list"
	if err := store.AppendActivityEntry("activity-1", "10.0.0.1", "api.example.com", "A", "blocked", &ruleID, &ruleSource, &responseTime); err != nil {
		t.Fatalf("AppendActivityEntry() first error = %v", err)
	}
	if err := store.AppendActivityEntry("activity-2", "10.0.0.1", "api.example.com", "A", "blocked", &ruleID, &ruleSource, &responseTime); err != nil {
		t.Fatalf("AppendActivityEntry() second error = %v", err)
	}
	if err := store.AppendActivityEntry("activity-3", "10.0.0.2", "www.example.com", "A", "allowed", nil, nil, &responseTime); err != nil {
		t.Fatalf("AppendActivityEntry() third error = %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity?client_ip=10.0.0.1&action=blocked&limit=1&offset=1", nil)
	handler.ListActivity(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ListActivity() status = %d, want %d", rec.Code, http.StatusOK)
	}
	var payload struct {
		Total int             `json:"total"`
		Items []activityEntry `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decoding ListActivity() response: %v", err)
	}
	if payload.Total != 2 {
		t.Fatalf("ListActivity() total = %d, want %d", payload.Total, 2)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("ListActivity() items len = %d, want %d", len(payload.Items), 1)
	}
	if payload.Items[0].ClientIP != "10.0.0.1" || payload.Items[0].Action != "blocked" {
		t.Fatalf("ListActivity() item = %+v, want filtered blocked entry", payload.Items[0])
	}
}

func openBoltActivityHandlerStore(t *testing.T) *storage.BoltStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openfiltr.db")
	store, err := storage.OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() error = %v", err)
	}
	return store
}
