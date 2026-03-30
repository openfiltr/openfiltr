package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	bboltStoreVersion   = 1
	bboltMetadataBucket = "metadata"
	bboltVersionKey     = "store_version"
)

var bboltRequiredBuckets = []string{
	"allow_rules",
	"allow_rules_lookup",
	"api_tokens",
	"activity_log",
	"audit_events",
	"block_rules",
	"block_rules_lookup",
	"clients",
	"dns_entries",
	"dns_entries_lookup",
	"groups",
	"rule_sources",
	"upstream_servers",
	"users",
	"users_by_username",
}

// BoltStore bootstraps and wraps the bbolt database file.
type BoltStore struct {
	db   *bolt.DB
	shim *sql.DB
}

func OpenBolt(path string) (*BoltStore, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("opening bbolt database: %w", err)
	}

	shim, err := openSQLShim()
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("opening SQL shim: %w", err)
	}

	if err := bootstrapBolt(db); err != nil {
		_ = shim.Close()
		_ = db.Close()
		return nil, err
	}

	return &BoltStore{db: db, shim: shim}, nil
}

func (s *BoltStore) Close() error {
	if s.shim != nil {
		_ = s.shim.Close()
	}
	return s.db.Close()
}

func (s *BoltStore) Ping() error {
	if s.shim == nil {
		return nil
	}
	return s.shim.Ping()
}

func (s *BoltStore) Exec(query string, args ...any) (sql.Result, error) {
	return s.shim.Exec(query, args...)
}

func (s *BoltStore) Query(query string, args ...any) (*sql.Rows, error) {
	return s.shim.Query(query, args...)
}

func (s *BoltStore) QueryRow(query string, args ...any) *sql.Row {
	return s.shim.QueryRow(query, args...)
}

func (s *BoltStore) Begin() (*sql.Tx, error) {
	return s.shim.Begin()
}

func (s *BoltStore) View(fn func(*bolt.Tx) error) error {
	return s.db.View(fn)
}

func (s *BoltStore) Update(fn func(*bolt.Tx) error) error {
	return s.db.Update(fn)
}

func bootstrapBolt(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		meta, err := tx.CreateBucketIfNotExists([]byte(bboltMetadataBucket))
		if err != nil {
			return fmt.Errorf("creating metadata bucket: %w", err)
		}

		for _, name := range bboltRequiredBuckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
				return fmt.Errorf("creating bucket %q: %w", name, err)
			}
		}

		version := meta.Get([]byte(bboltVersionKey))
		switch {
		case version == nil:
			if err := meta.Put([]byte(bboltVersionKey), []byte(strconv.Itoa(bboltStoreVersion))); err != nil {
				return fmt.Errorf("writing metadata version: %w", err)
			}
		case string(version) != strconv.Itoa(bboltStoreVersion):
			return fmt.Errorf("unsupported bbolt store version %q: expected %d", version, bboltStoreVersion)
		}

		return nil
	})
}

const boltSQLShimDriverName = "openfiltr-bbolt-sql-shim"

var boltSQLShimOnce sync.Once

func openSQLShim() (*sql.DB, error) {
	boltSQLShimOnce.Do(func() {
		sql.Register(boltSQLShimDriverName, boltSQLShimDriver{})
	})
	return sql.Open(boltSQLShimDriverName, "")
}

type boltSQLShimDriver struct{}

func (boltSQLShimDriver) Open(string) (driver.Conn, error) {
	return boltSQLShimConn{}, nil
}

type boltSQLShimConn struct{}

func (boltSQLShimConn) Prepare(string) (driver.Stmt, error) {
	return nil, boltSQLShimErr
}

func (boltSQLShimConn) Close() error { return nil }

func (boltSQLShimConn) Begin() (driver.Tx, error) {
	return nil, boltSQLShimErr
}

func (boltSQLShimConn) Ping(context.Context) error { return boltSQLShimErr }

func (boltSQLShimConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return nil, boltSQLShimErr
}

func (boltSQLShimConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, boltSQLShimErr
}

var boltSQLShimErr = errors.New("bbolt backend does not support SQL queries yet")
