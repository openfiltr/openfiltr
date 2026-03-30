package dns

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/openfiltr/openfiltr/internal/storage"
)

func TestBoltServerUsesRuleIndexes(t *testing.T) {
	store := openBoltDNSStore(t)
	defer store.Close()

	comment := "rule"
	if _, err := store.CreateRule("block_rules", "rule-exact", "example.com", "exact", &comment, 1, nil); err != nil {
		t.Fatalf("CreateRule() exact error = %v", err)
	}
	if _, err := store.CreateRule("block_rules", "rule-wild", "*.example.com", "wildcard", &comment, 1, nil); err != nil {
		t.Fatalf("CreateRule() wildcard error = %v", err)
	}
	if _, err := store.CreateRule("block_rules", "rule-regex", `^api\.example\.com$`, "regex", &comment, 1, nil); err != nil {
		t.Fatalf("CreateRule() regex error = %v", err)
	}

	srv := NewServer(nil, store)
	if !srv.hasExactBlockRule("example.com") {
		t.Fatal("hasExactBlockRule() = false, want true")
	}
	if !srv.hasWildcardBlockRule("foo.example.com") {
		t.Fatal("hasWildcardBlockRule() = false, want true")
	}
	if !srv.hasRegexBlockRule("API.Example.Com") {
		t.Fatal("hasRegexBlockRule() = false, want true")
	}
}

func TestBoltServerLocalEntriesUsesHostAndTypeIndex(t *testing.T) {
	store := openBoltDNSStore(t)
	defer store.Close()

	comment := "dns"
	createdBy := "user-1"
	if _, err := store.CreateDNSEntry("dns-1", "example.com", "A", "1.1.1.1", 300, &comment, &createdBy, 1); err != nil {
		t.Fatalf("CreateDNSEntry() first error = %v", err)
	}
	if _, err := store.CreateDNSEntry("dns-2", "example.com", "A", "1.1.1.2", 120, &comment, &createdBy, 1); err != nil {
		t.Fatalf("CreateDNSEntry() second error = %v", err)
	}
	if _, err := store.CreateDNSEntry("dns-3", "example.com", "AAAA", "2001:db8::1", 300, &comment, &createdBy, 1); err != nil {
		t.Fatalf("CreateDNSEntry() third error = %v", err)
	}
	if _, err := store.CreateDNSEntry("dns-4", "example.com", "A", "1.1.1.3", 60, &comment, &createdBy, 0); err != nil {
		t.Fatalf("CreateDNSEntry() disabled error = %v", err)
	}

	srv := NewServer(nil, store)
	rrs := srv.localEntries("example.com", dns.TypeA)
	if len(rrs) != 2 {
		t.Fatalf("localEntries() len = %d, want %d", len(rrs), 2)
	}
	rrs = srv.localEntries("example.com", dns.TypeAAAA)
	if len(rrs) != 1 {
		t.Fatalf("localEntries() AAAA len = %d, want %d", len(rrs), 1)
	}
}

func openBoltDNSStore(t *testing.T) *storage.BoltStore {
	t.Helper()
	path := t.TempDir() + "/openfiltr.db"
	store, err := storage.OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() error = %v", err)
	}
	return store
}
