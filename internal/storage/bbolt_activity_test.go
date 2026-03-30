package storage

import (
	"path/filepath"
	"testing"
)

func TestBoltActivityLogQueries(t *testing.T) {
	store := openBoltActivityStore(t)
	defer store.Close()

	ruleID := "rule-1"
	ruleSource := "block-list"
	responseTime := 12

	if err := store.AppendActivityEntry("activity-1", "10.0.0.1", "api.example.com", "A", "blocked", &ruleID, &ruleSource, &responseTime); err != nil {
		t.Fatalf("AppendActivityEntry() first error = %v", err)
	}
	if err := store.AppendActivityEntry("activity-2", "10.0.0.1", "api.example.com", "A", "blocked", &ruleID, &ruleSource, &responseTime); err != nil {
		t.Fatalf("AppendActivityEntry() second error = %v", err)
	}
	if err := store.AppendActivityEntry("activity-3", "10.0.0.2", "www.example.com", "A", "allowed", nil, nil, &responseTime); err != nil {
		t.Fatalf("AppendActivityEntry() third error = %v", err)
	}

	items, total, err := store.ListActivity(100, 0, "10.0.0.1", "", "blocked")
	if err != nil {
		t.Fatalf("ListActivity() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("ListActivity() total = %d, want %d", total, 2)
	}
	if len(items) != 2 {
		t.Fatalf("ListActivity() len = %d, want %d", len(items), 2)
	}

	totalCount, blocked, allowed, err := store.ActivityCounts()
	if err != nil {
		t.Fatalf("ActivityCounts() error = %v", err)
	}
	if totalCount != 3 || blocked != 2 || allowed != 1 {
		t.Fatalf("ActivityCounts() = (%d,%d,%d), want (3,2,1)", totalCount, blocked, allowed)
	}

	top, err := store.TopBlockedDomains(10)
	if err != nil {
		t.Fatalf("TopBlockedDomains() error = %v", err)
	}
	if len(top) != 1 {
		t.Fatalf("TopBlockedDomains() len = %d, want %d", len(top), 1)
	}
	if top[0].Domain != "api.example.com" || top[0].Count != 2 {
		t.Fatalf("TopBlockedDomains()[0] = %#v, want domain %q count %d", top[0], "api.example.com", 2)
	}
}

func openBoltActivityStore(t *testing.T) *BoltStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openfiltr.db")
	store, err := OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() error = %v", err)
	}
	return store
}
