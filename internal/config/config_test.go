package config

import (
	"path/filepath"
	"testing"
)

func TestLoadWritesBoltDefaults(t *testing.T) {
	t.Setenv("OPENFILTR_JWT_SECRET", "")
	t.Setenv("OPENFILTR_DATABASE_URL", "")
	t.Setenv("OPENFILTR_DATABASE_PATH", "")

	cfgPath := filepath.Join(t.TempDir(), "config", "app.yaml")
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Storage.DatabasePath != DefaultDatabasePath {
		t.Fatalf("Storage.DatabasePath = %q, want %q", cfg.Storage.DatabasePath, DefaultDatabasePath)
	}
	if cfg.Storage.DatabaseURL != "" {
		t.Fatalf("Storage.DatabaseURL = %q, want empty", cfg.Storage.DatabaseURL)
	}
}

func TestLoadHonoursDatabaseURLOverride(t *testing.T) {
	t.Setenv("OPENFILTR_DATABASE_URL", "postgres://example:secret@db:5432/openfiltr?sslmode=disable")
	t.Setenv("OPENFILTR_DATABASE_PATH", "")

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

func TestLoadHonoursDatabasePathOverride(t *testing.T) {
	t.Setenv("OPENFILTR_DATABASE_URL", "")
	t.Setenv("OPENFILTR_DATABASE_PATH", "/var/lib/openfiltr/openfiltr.db")

	cfgPath := filepath.Join(t.TempDir(), "config", "app.yaml")
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := "/var/lib/openfiltr/openfiltr.db"
	if cfg.Storage.DatabasePath != want {
		t.Fatalf("Storage.DatabasePath = %q, want %q", cfg.Storage.DatabasePath, want)
	}
}
