package storage

import (
	"path/filepath"
	"testing"
)

func TestBoltStoreRuleIndexesAndCrud(t *testing.T) {
	store := openBoltCatalogStore(t)
	defer store.Close()

	exactComment := "exact rule"
	exactCreatedBy := "user-1"
	exact, err := store.CreateRule("block_rules", "rule-exact", "Example.COM", "exact", &exactComment, 1, &exactCreatedBy)
	if err != nil {
		t.Fatalf("CreateRule() exact error = %v", err)
	}
	wildcardComment := "wildcard rule"
	wildcard, err := store.CreateRule("block_rules", "rule-wildcard", "*.example.com", "wildcard", &wildcardComment, 1, nil)
	if err != nil {
		t.Fatalf("CreateRule() wildcard error = %v", err)
	}
	regexComment := "regex rule"
	if _, err := store.CreateRule("block_rules", "rule-regex", `^api\.example\.com$`, "regex", &regexComment, 1, nil); err != nil {
		t.Fatalf("CreateRule() regex error = %v", err)
	}
	allowComment := "allow rule"
	if _, err := store.CreateRule("allow_rules", "rule-allow", "allowed.example.com", "exact", &allowComment, 1, nil); err != nil {
		t.Fatalf("CreateRule() allow error = %v", err)
	}

	if exists, err := store.HasRulePattern("block_rules", "exact", "example.com"); err != nil {
		t.Fatalf("HasRulePattern() exact error = %v", err)
	} else if !exists {
		t.Fatal("HasRulePattern() exact = false, want true")
	}
	if exists, err := store.HasRulePattern("block_rules", "wildcard", "*.example.com"); err != nil {
		t.Fatalf("HasRulePattern() wildcard error = %v", err)
	} else if !exists {
		t.Fatal("HasRulePattern() wildcard = false, want true")
	}
	if exists, err := store.HasRulePattern("allow_rules", "exact", "allowed.example.com"); err != nil {
		t.Fatalf("HasRulePattern() allow error = %v", err)
	} else if !exists {
		t.Fatal("HasRulePattern() allow = false, want true")
	}

	patterns, err := store.ListRulePatternsByType("block_rules", "regex")
	if err != nil {
		t.Fatalf("ListRulePatternsByType() error = %v", err)
	}
	if len(patterns) != 1 || patterns[0] != `^api\.example\.com$` {
		t.Fatalf("ListRulePatternsByType() = %v, want [%q]", patterns, `^api\.example\.com$`)
	}

	items, total, err := store.ListRules("block_rules", 100, 0)
	if err != nil {
		t.Fatalf("ListRules() error = %v", err)
	}
	if total != 3 {
		t.Fatalf("ListRules() total = %d, want %d", total, 3)
	}
	if len(items) != 3 {
		t.Fatalf("ListRules() len = %d, want %d", len(items), 3)
	}

	disabled := 0
	updated, found, err := store.UpdateRule("block_rules", exact.ID, nil, nil, nil, &disabled)
	if err != nil {
		t.Fatalf("UpdateRule() disable error = %v", err)
	}
	if !found {
		t.Fatal("UpdateRule() disable not found")
	}
	if updated.Enabled != 0 {
		t.Fatalf("UpdateRule() enabled = %d, want 0", updated.Enabled)
	}
	if exists, err := store.HasRulePattern("block_rules", "exact", "example.com"); err != nil {
		t.Fatalf("HasRulePattern() exact after disable error = %v", err)
	} else if exists {
		t.Fatal("HasRulePattern() exact after disable = true, want false")
	}

	deleted, err := store.DeleteRule("block_rules", wildcard.ID)
	if err != nil {
		t.Fatalf("DeleteRule() error = %v", err)
	}
	if !deleted {
		t.Fatal("DeleteRule() returned false")
	}
	if exists, err := store.HasRulePattern("block_rules", "wildcard", "*.example.com"); err != nil {
		t.Fatalf("HasRulePattern() wildcard after delete error = %v", err)
	} else if exists {
		t.Fatal("HasRulePattern() wildcard after delete = true, want false")
	}
}

func TestBoltStoreDNSEntryIndexesAndCrud(t *testing.T) {
	store := openBoltCatalogStore(t)
	defer store.Close()

	createdBy := "user-1"
	comment := "first"
	first, err := store.CreateDNSEntry("dns-1", "Example.com", "A", "1.1.1.1", 300, &comment, &createdBy, 1)
	if err != nil {
		t.Fatalf("CreateDNSEntry() first error = %v", err)
	}
	secondComment := "second"
	second, err := store.CreateDNSEntry("dns-2", "Example.com", "A", "1.1.1.2", 120, &secondComment, &createdBy, 1)
	if err != nil {
		t.Fatalf("CreateDNSEntry() second error = %v", err)
	}
	disabledComment := "disabled"
	if _, err := store.CreateDNSEntry("dns-3", "Example.com", "A", "1.1.1.3", 60, &disabledComment, &createdBy, 0); err != nil {
		t.Fatalf("CreateDNSEntry() disabled error = %v", err)
	}

	items, total, err := store.ListDNSEntries(100, 0)
	if err != nil {
		t.Fatalf("ListDNSEntries() error = %v", err)
	}
	if total != 3 {
		t.Fatalf("ListDNSEntries() total = %d, want %d", total, 3)
	}
	if len(items) != 3 {
		t.Fatalf("ListDNSEntries() len = %d, want %d", len(items), 3)
	}

	lookups, err := store.DNSEntriesByHostAndType("example.com", "A")
	if err != nil {
		t.Fatalf("DNSEntriesByHostAndType() error = %v", err)
	}
	if len(lookups) != 2 {
		t.Fatalf("DNSEntriesByHostAndType() len = %d, want %d", len(lookups), 2)
	}

	newHost := "www.example.com"
	updated, found, err := store.UpdateDNSEntry(second.ID, &newHost, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("UpdateDNSEntry() error = %v", err)
	}
	if !found {
		t.Fatal("UpdateDNSEntry() not found")
	}
	if updated.Host != newHost {
		t.Fatalf("UpdateDNSEntry() host = %q, want %q", updated.Host, newHost)
	}

	lookups, err = store.DNSEntriesByHostAndType("example.com", "A")
	if err != nil {
		t.Fatalf("DNSEntriesByHostAndType() after update error = %v", err)
	}
	if len(lookups) != 1 {
		t.Fatalf("DNSEntriesByHostAndType() after update len = %d, want %d", len(lookups), 1)
	}
	lookups, err = store.DNSEntriesByHostAndType("www.example.com", "A")
	if err != nil {
		t.Fatalf("DNSEntriesByHostAndType() moved entry error = %v", err)
	}
	if len(lookups) != 1 {
		t.Fatalf("DNSEntriesByHostAndType() moved entry len = %d, want %d", len(lookups), 1)
	}

	deleted, err := store.DeleteDNSEntry(first.ID)
	if err != nil {
		t.Fatalf("DeleteDNSEntry() error = %v", err)
	}
	if !deleted {
		t.Fatal("DeleteDNSEntry() returned false")
	}
	if _, err := store.GetDNSEntry(first.ID); err == nil {
		t.Fatal("GetDNSEntry() after delete = nil error, want not found")
	}
}

func openBoltCatalogStore(t *testing.T) *BoltStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openfiltr.db")
	store, err := OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() error = %v", err)
	}
	return store
}
