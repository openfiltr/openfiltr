package main

import (
	"path/filepath"
	"testing"

	"github.com/openfiltr/openfiltr/internal/config"
)

func TestResolveStorageSelectionDefaultsToBolt(t *testing.T) {
	cfg := config.Defaults()
	cfg.Storage.DatabaseURL = ""
	cfg.Storage.DatabasePath = ""

	configPath := filepath.Join(t.TempDir(), "config", "app.yaml")
	selection := resolveStorageSelection(cfg, configPath)

	if selection.backend != "bolt" {
		t.Fatalf("backend = %q, want %q", selection.backend, "bolt")
	}
	want := filepath.Join(filepath.Dir(configPath), config.DefaultDatabasePath)
	if selection.path != want {
		t.Fatalf("path = %q, want %q", selection.path, want)
	}
}

func TestResolveStorageSelectionUsesPostgresWhenURLSet(t *testing.T) {
	cfg := config.Defaults()
	cfg.Storage.DatabaseURL = "postgres://example:secret@db:5432/openfiltr?sslmode=disable"
	cfg.Storage.DatabasePath = "custom.db"

	selection := resolveStorageSelection(cfg, filepath.Join(t.TempDir(), "config", "app.yaml"))

	if selection.backend != "postgres" {
		t.Fatalf("backend = %q, want %q", selection.backend, "postgres")
	}
	if selection.path != "" {
		t.Fatalf("path = %q, want empty", selection.path)
	}
}
