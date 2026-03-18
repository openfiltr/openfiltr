package storage

import "testing"

func TestRebind(t *testing.T) {
	query := "SELECT id FROM users WHERE username=? AND role=? LIMIT ? OFFSET ?"
	got := Rebind(query)
	want := "SELECT id FROM users WHERE username=$1 AND role=$2 LIMIT $3 OFFSET $4"
	if got != want {
		t.Fatalf("Rebind() = %q, want %q", got, want)
	}
}

func TestMigrationVersion(t *testing.T) {
	got, err := migrationVersion("migrations/001_initial_schema.sql")
	if err != nil {
		t.Fatalf("migrationVersion() error = %v", err)
	}
	if got != 1 {
		t.Fatalf("migrationVersion() = %d, want 1", got)
	}
}
