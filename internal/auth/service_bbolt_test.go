package auth

import (
	"path/filepath"
	"testing"

	"github.com/openfiltr/openfiltr/internal/storage"
)

func TestBoltAuthServicePersistsUserAcrossReopen(t *testing.T) {
	store, path := openBoltAuthStore(t)

	svc := NewService(store, "secret", 1)
	if err := svc.CreateAdminUser("alice", "correct horse battery staple"); err != nil {
		t.Fatalf("CreateAdminUser() error = %v", err)
	}
	if count, err := svc.CountUsers(); err != nil {
		t.Fatalf("CountUsers() error = %v", err)
	} else if count != 1 {
		t.Fatalf("CountUsers() = %d, want %d", count, 1)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reopened, err := storage.OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() reopen error = %v", err)
	}
	defer reopened.Close()

	svc = NewService(reopened, "secret", 1)
	user, err := svc.LookupUserByUsername("alice")
	if err != nil {
		t.Fatalf("LookupUserByUsername() error = %v", err)
	}
	if user.Username != "alice" {
		t.Fatalf("Username = %q, want %q", user.Username, "alice")
	}
	if user.Role != "admin" {
		t.Fatalf("Role = %q, want %q", user.Role, "admin")
	}
	if !CheckPassword("correct horse battery staple", user.PasswordHash) {
		t.Fatal("stored password hash no longer matches the original password")
	}
}

func TestBoltAuthServicePersistsAndValidatesAPITokens(t *testing.T) {
	store, path := openBoltAuthStore(t)

	svc := NewService(store, "secret", 1)
	if err := svc.CreateAdminUser("alice", "correct horse battery staple"); err != nil {
		t.Fatalf("CreateAdminUser() error = %v", err)
	}
	user, err := svc.LookupUserByUsername("alice")
	if err != nil {
		t.Fatalf("LookupUserByUsername() error = %v", err)
	}

	raw, hash, err := GenerateAPIToken()
	if err != nil {
		t.Fatalf("GenerateAPIToken() error = %v", err)
	}
	tokenID, err := svc.CreateAPIToken(user.ID, "cli", hash, nil)
	if err != nil {
		t.Fatalf("CreateAPIToken() error = %v", err)
	}
	if tokenID == "" {
		t.Fatal("CreateAPIToken() returned an empty id")
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reopened, err := storage.OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() reopen error = %v", err)
	}
	defer reopened.Close()

	svc = NewService(reopened, "secret", 1)
	claims, err := svc.ValidateAPIToken(raw)
	if err != nil {
		t.Fatalf("ValidateAPIToken() error = %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("Claims.UserID = %q, want %q", claims.UserID, user.ID)
	}
	if claims.Username != "alice" {
		t.Fatalf("Claims.Username = %q, want %q", claims.Username, "alice")
	}
	if claims.Role != "admin" {
		t.Fatalf("Claims.Role = %q, want %q", claims.Role, "admin")
	}

	items, err := svc.ListAPITokens(user.ID)
	if err != nil {
		t.Fatalf("ListAPITokens() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListAPITokens() len = %d, want %d", len(items), 1)
	}
	if items[0].ID != tokenID {
		t.Fatalf("ListAPITokens()[0].ID = %q, want %q", items[0].ID, tokenID)
	}
	if items[0].LastUsedAt == nil {
		t.Fatal("ValidateAPIToken() did not update last_used_at")
	}

	deleted, err := svc.DeleteAPIToken(tokenID, user.ID)
	if err != nil {
		t.Fatalf("DeleteAPIToken() error = %v", err)
	}
	if !deleted {
		t.Fatal("DeleteAPIToken() returned false")
	}

	items, err = svc.ListAPITokens(user.ID)
	if err != nil {
		t.Fatalf("ListAPITokens() after delete error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("ListAPITokens() after delete len = %d, want %d", len(items), 0)
	}
}

func openBoltAuthStore(t *testing.T) (*storage.BoltStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openfiltr.db")
	store, err := storage.OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() error = %v", err)
	}
	return store, path
}
