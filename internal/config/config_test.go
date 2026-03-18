package config

import (
	"path/filepath"
	"testing"
)

func TestLoadWritesPostgresDefaults(t *testing.T) {
	t.Setenv("OPENFILTR_JWT_SECRET", "")
	t.Setenv("OPENFILTR_DATABASE_URL", "")

	cfgPath := filepath.Join(t.TempDir(), "config", "app.yaml")
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := "postgres://openfiltr:openfiltr@localhost:5432/openfiltr?sslmode=disable"
	if cfg.Storage.DatabaseURL != want {
		t.Fatalf("Storage.DatabaseURL = %q, want %q", cfg.Storage.DatabaseURL, want)
	}
}

func TestLoadHonoursDatabaseURLOverride(t *testing.T) {
	t.Setenv("OPENFILTR_DATABASE_URL", "postgres://example:secret@db:5432/openfiltr?sslmode=disable")

	cfgPath := filepath.Join(t.TempDir(), "config", "app.yaml")
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := "postgres://example:secret@db:5432/openfiltr?sslmode=disable"
	if cfg.Storage.DatabaseURL != want {
		t.Fatalf("Storage.DatabaseURL = %q, want %q", cfg.Storage.DatabaseURL, want)
	}
}
