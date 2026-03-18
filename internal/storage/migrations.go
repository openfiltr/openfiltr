package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	files, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("listing migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		version, err := migrationVersion(file)
		if err != nil {
			return err
		}

		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version=$1", version).Scan(&count); err != nil {
			return fmt.Errorf("checking migration %d: %w", version, err)
		}
		if count > 0 {
			continue
		}

		sqlBytes, err := migrationFiles.ReadFile(file)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", file, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("starting migration %d: %w", version, err)
		}
		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("running migration %d: %w", version, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations(version) VALUES($1)", version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %d: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", version, err)
		}
	}

	return nil
}

func migrationVersion(path string) (int, error) {
	name := filepath.Base(path)
	prefix, _, found := strings.Cut(name, "_")
	if !found {
		return 0, fmt.Errorf("invalid migration name %q", name)
	}
	version, err := strconv.Atoi(prefix)
	if err != nil {
		return 0, fmt.Errorf("invalid migration version %q: %w", name, err)
	}
	return version, nil
}
